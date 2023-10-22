package media

import (
	"fmt"
	"net/http"
)

func Media(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
    case "GET":
		// comic update
		fmt.Println("TEST")
	case "POST":
		// video update
		// video_update_log := VidPrevImgUpt(w, r)
		fmt.Println("TEST")
	case "PUT":
		// novel update
		fmt.Println("TEST")
	default:
		fmt.Println("TEST")
	}
}