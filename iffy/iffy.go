package iffy

func PanicIf(err error) {
  if err != nil {
    panic(err)
  }
}

func Disregard(err error) {
}
