package buildinfo

import "log"

var (
	GitVersion   string
	BuildTime    string
	Architecture string
	GoVersion    string
)

func Print() {
	log.Println("[Build] Git Version:", GitVersion)
	log.Println("[Build] Build Time:", BuildTime)
	log.Println("[Build] Architecture:", Architecture)
	log.Println("[Build] Go Version:", GoVersion)
}
