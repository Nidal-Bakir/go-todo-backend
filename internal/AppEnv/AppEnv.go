package AppEnv

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var (
	isProd  = false
	isStag  = false
	isLocal = false
	EnvName = ""
)

func init() {
	file, err := os.Open(".env")
	if err != nil {
		pwd, _ := os.Getwd()
		log.Fatal("error: can not read .env file, pwd= ", pwd)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, val, ok := strings.Cut(line, "=")
		if !ok || line[0] == '#' {
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
		isLocal = true
	case "stag":
		isStag = true
	case "prod":
		isProd = true
	default:
		log.Fatal("The value for APP_ENV in the .env file not determined, aborting...")
	}

	EnvName = appEnv
}

func IsProd() bool {
	return isProd
}
func IsStag() bool {
	return isStag
}
func IsLocal() bool {
	return isLocal
}

func IsStagOrLocal() bool {
	return IsStag() || IsLocal()
}
