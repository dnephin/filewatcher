package runner

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-tty"
)

// Update contains values to update in the main event loop
type Update struct {
	Env string
}

func scanInput(chUpdate chan Update) {
	tty, err := tty.Open()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer tty.Close()
	for {
		r, err := tty.ReadRune()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		switch r {
		case 'e':
			env, err := getEnvVar(tty)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			chUpdate <- Update{Env: env}
		}
	}
}

func getEnvVar(tty *tty.TTY) (string, error) {
	var r rune
	var env string

	for {
		var err error
		r, err = tty.ReadRune()
		switch {
		case err != nil:
			return "", err
		case r == rune(13):
			// TODO: check for valid env var
			fmt.Println("\nUpdated")
			return env, nil
		}
		s := string(r)
		env += s
		fmt.Print(s)
	}
}
