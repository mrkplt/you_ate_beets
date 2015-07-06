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

//TODO - Move BeetTweet persistance, sanitation and definition logic out.
package main

import (
	"time"
	"github.com/mrkplt/you_ate_beets/beet_tweet"
)

func main() {
	api := beetTweet.SetupApi()
	db := beetTweet.SetupDB()
	//TODO - Now that you know this works, reimplement channeling.
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		ats := beetTweet.GetMentions(api, db)

		for _, t := range ats {
			tweet := beetTweet.ProcessTweet(t)
			beetTweet.PersistTweet(tweet, db)
		}

		bts := beetTweet.RetrieveTweets(db)

		beetTweet.PostTweets(api, bts, db)
	}
}
