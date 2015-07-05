// CREATE DATABASE you_ate_beets;
// CREATE USER you_ate_beets PASSWORD 'your_password';
// CREATE TABLE tweet(
//    ID                  BIGINT      PRIMARY KEY         NOT NULL,
//    NAME                CHAR(20)                        NOT NULL,
//    SCREENNAME          CHAR(15)                        NOT NULL,
//    HOURS               INT                             NOT NULL,
//    MENTIONTIME         TIMESTAMP WITHOUT TIME ZONE     NOT NULL,
//    NOTIFICATIONTIME    TIMESTAMP WITHOUT TIME ZONE     NOT NULL,
//    NOTIFIED            BOOL                            NOT NULL);

// \c you_ate_beets you_ate_beets
//SELECT now()::date, now()::time;

//TODO - Move beettweet persistance, sanitation and definition logic out.
package main

import (
	"database/sql"
	"fmt"
	"github.com/chimeracoder/anaconda"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"
	"net/url"
	"regexp"
	"github.com/mrkplt/you_ate_beets/config"
	"github.com/mrkplt/you_ate_beets/iffy"
)

type beetTweet struct {
	Id               int64
	ScreenName       string
	Name             string
	Hours            int
	MentionTime      time.Time
	NotificationTime time.Time
	Notified         bool
}

func main() {
	api := setupApi()
	db := setupDB()
	//TODO - Now that you know this works, reimplement channeling.
	ticker := time.NewTicker(1	 * time.Minute)
	for range ticker.C {
		ats := getMentions(api, db)

		for _, t := range ats {
			tweet := processTweet(t)
			if tweet.Hours != -1 {
				persistTweet(tweet, db)
			}
		}

		bts := retrieveTweets(db)

		postTweets(api, bts, db)
	}
}

func getMentions(api *anaconda.TwitterApi, db *sql.DB) (retrieved_tweets []anaconda.Tweet) {
	var lastMention int64
	v := url.Values{}
	v.Add("count", "200")
	v.Add("include_rts", "1")
	err := db.QueryRow("SELECT ID FROM tweet ORDER BY ID DESC LIMIT 1;").Scan(&lastMention)

	if err == nil {
		lastMentionStr := strconv.FormatInt(lastMention, 10)
		v.Add("since_id", lastMentionStr)
	}

	retrieved_tweets, err = api.GetMentionsTimeline(v)
	iffy.PanicIf(err)
	return
}

func processTweet(tweet anaconda.Tweet) (bt *beetTweet) {
	regex, err := regexp.Compile("\\d+")
	iffy.Disregard(err)
	numbers := regex.FindAllString(tweet.Text, 70)
	if numbers == nil {
		numbers = append(numbers, "-1")
	}

	numberString := numbers[len(numbers)-1]

	hours, err := strconv.Atoi(numberString)
	iffy.Disregard(err)
	mentionTime, err := time.Parse(time.RubyDate, tweet.CreatedAt)
	iffy.Disregard(err)
	notificationTime := mentionTime.Add(time.Duration(hours) * time.Hour)

	bt = &beetTweet{
		Id:               tweet.Id,
		ScreenName:       tweet.User.ScreenName,
		Name:             tweet.User.Name,
		Hours:            hours,
		MentionTime:      mentionTime,
		NotificationTime: notificationTime,
		Notified:         false,
	}
	return
}

func persistTweet(bt *beetTweet, db *sql.DB) {
	query := fmt.Sprintf("INSERT INTO tweet (ID, SCREENNAME, NAME, HOURS, MENTIONTIME, NOTIFICATIONTIME, NOTIFIED) VALUES (%d, '%s', '%s', %d, '%s', '%s', %t);",
		bt.Id,
		bt.ScreenName,
		bt.Name,
		bt.Hours,
		bt.MentionTime.Format(time.RFC3339),
		bt.NotificationTime.Format(time.RFC3339),
		bt.Notified)

	_, err := db.Query(query)
	if err != nil {
		if _, ok := err.(*pq.Error); ok {
			return
		}
	}
	return
}

func retrieveTweets(db *sql.DB) (beetTweets []beetTweet) {
	query := fmt.Sprintf("SELECT * FROM tweet where NOTIFIED = false AND NOTIFICATIONTIME <= '%s';", time.Now().UTC().Format(time.RFC3339))

	rows, err := db.Query(query)

	beetTweets = []beetTweet{}

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		bt := beetTweet{}

		err = rows.Scan(&bt.Id,
			&bt.ScreenName,
			&bt.Name,
			&bt.Hours,
			&bt.MentionTime,
			&bt.NotificationTime,
			&bt.Notified)

		iffy.PanicIf(err)

		bt.Name = strings.TrimSpace(bt.Name)
		bt.ScreenName = strings.TrimSpace(bt.ScreenName)
		beetTweets = append(beetTweets, bt)
	}
	return
}

func postTweets(api *anaconda.TwitterApi, bts []beetTweet, db *sql.DB) {
	for _, bt := range bts {
		//TODO: Psuedo random reminder messages. For fun, and not angering the twitter api.
		_, err := api.PostTweet(fmt.Sprintf("@%s The number is %d.", bt.Name, bt.Hours), nil)

		if err == nil {
			query := fmt.Sprintf("UPDATE tweet SET NOTIFIED = true WHERE ID = %d;", bt.Id)
			db.Query(query)
		}
	}
}

func secrets() config.Config {
	return config.Secrets()
}

func setupApi() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(secrets().Anaconda.ConsumerKey)
	anaconda.SetConsumerSecret(secrets().Anaconda.ConsumerSecret)
	api := anaconda.NewTwitterApi(secrets().Anaconda.AccessToken, secrets().Anaconda.AccessSecret)
	return api
}

func connectionString() string {
	return fmt.Sprintf("dbname=%s sslmode=disable", secrets().Database.Name)
}

func setupDB() *sql.DB {
	db, err := sql.Open("postgres", connectionString())
	iffy.PanicIf(err)

	return db
}
