package AppEnv

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var (
	IsProd  = false
	IsStag  = false
	IsLocal = false
)

func init() {
	file, err := os.OpenFile(".env", os.O_RDONLY, os.ModeAppend)
	if err != nil {
		log.Fatal("can not read .env file")
	}
	defer file.Close()

	v := bufio.NewScanner(file)

	for v.Scan() {
		key, val, ok := strings.Cut(v.Text(), "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		err := os.Setenv(key, val)
		if err != nil {
			log.Fatal("can not set env var error: ", err)
		}
	}

	appEnv := os.Getenv("APP_ENV")
	switch appEnv {
	case "local":
		IsLocal = true
	case "stag":
		IsStag = true
	case "prod":
		IsProd = true
	default:
		log.Fatal("The value for APP_ENV in the .env file not determined, aborting...")
	}
}
