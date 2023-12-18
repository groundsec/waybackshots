package utils

import "fmt"

func Banner(version string) {
	banner := `                           
               _           _       _       _       
 _ _ _ ___ _ _| |_ ___ ___| |_ ___| |_ ___| |_ ___ 
| | | | .'| | | . | .'|  _| '_|_ -|   | . |  _|_ -|
|_____|__,|_  |___|__,|___|_,_|___|_|_|___|_| |___|
          |___|                                    
v%s - https://github.com/groundsec/waybackshots

`
	fmt.Printf(banner, version)
}
