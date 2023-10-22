package comic

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	// "os/exec"
	"log"
	"strings"
	"strconv"
	"encoding/json"
	"database/sql"

	"example.com/m/auth"
	"example.com/m/db"
)

type ComicInfo struct {
	Title			string
	PrevImgPath		string
	Tag				string
	Translated		int
	Author			string
	Rating			int
}

type ComicInfoList struct {
	ComicInfo	[]ComicInfo
	TotalCount	int
}

type ComicPage struct {
	Id int
	ImgPath string
}

type ComicPart struct {
	Subtitle	string
	Pages	[]ComicPage
}

type ComicDetails struct {
	ComicInfo
	Parts	[]ComicPart
}

type RequestComicDetails struct {
	ComicName	string
}

type RequestComicUpdate struct {
	Action	string // refresh or edit
	ComicInfo
}

type Response struct {
	Status			int
}

func Comic(w http.ResponseWriter, r *http.Request) {
	if auth.VerifyPermission(r) {
		switch r.Method {
		case "GET":
			// transfer list
			comicListLog := ComicList(w, r)
			fmt.Println(comicListLog)
		case "POST":
			// transfer comic
			comicContentLog := ComicContent(w, r)
			fmt.Println(comicContentLog)
		case "PUT":
			// comic update
			comicUpdateLog := ComicUpdate(w, r)
			fmt.Println(comicUpdateLog)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

func ComicList(w http.ResponseWriter, r *http.Request) string {
	r.ParseForm()
	encodedTag := r.FormValue("tag")
	pageNumStr := r.FormValue("page")
	count, conditionString := ComicCountQuery(encodedTag)
	whereClause := ""
	if conditionString != "" {
		whereClause = strings.Join([]string{"WHERE tag LIKE", conditionString}, " ")
	}

	limitClause := "LIMIT 0, 18"
	pageNum, toIntErr := strconv.Atoi(pageNumStr)
	if toIntErr != nil {
		pageNum = 0
	}
	if pageNumStr != "" && pageNum > 0 && count > (pageNum - 1) * 18 {
		startNum := strconv.Itoa((pageNum - 1) * 18)
		limitClause = strings.Join([]string{"LIMIT ", startNum, ", 18"}, "")
	} // 1 - 0 ~ 17, 2 - 18 ~ 35, 3 - 36 ~ 53

	ComicQuery := "SELECT title, tag, translated, author, rating FROM comic"
	selectComicQuery := strings.Join([]string{ComicQuery, whereClause, "ORDER BY id DESC", limitClause}, " ")

	var comics ComicInfoList
	comics.TotalCount = count

	comicRows, queryErr := db.DB().Query(selectComicQuery)
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	defer comicRows.Close()

	hasRows := false

	for comicRows.Next() {
		hasRows = true
		var comic ComicInfo
		scanErr := comicRows.Scan(&comic.Title, &comic.Tag, &comic.Translated, &comic.Author, &comic.Rating)
		if scanErr != nil {
			log.Fatal(scanErr)
		}
		comicPath := strings.Join([]string{"/media/comic/", comic.Title, "/"}, "")
		comic.PrevImgPath = ComicPrevImg(comicPath)
		comics.ComicInfo = append(comics.ComicInfo, comic)
	}
	if !hasRows {
		return "Comic rows not found"
	}

	json.NewEncoder(w).Encode(comics)
	return "Comic List Response"
}

func ComicCountQuery(encodedTag string) (int, string) {
	var count int
	comicCountQuery := []string{}
	comicCountQuery = append(comicCountQuery, "SELECT COUNT(*) FROM comic")
	whereClause := ""

	if encodedTag != "" {
		decodedTag, decErr := url.QueryUnescape(encodedTag)
		if decErr != nil {
			return 0, whereClause
		}
		tags := strings.Split(decodedTag, " ")
		tagsWithWildcard := []string{}

		for _, tag := range tags {
			withWildcard := []string{"'%", tag, "%'"}
			tagsWithWildcard = append(tagsWithWildcard, strings.Join(withWildcard, ""))
		}
		whereClause = strings.Join(tagsWithWildcard, " AND tag LIKE ")
		comicCountQuery = append(comicCountQuery, "WHERE tag LIKE")
		comicCountQuery = append(comicCountQuery, whereClause)
	}
	selectComicCountQuery := strings.Join(comicCountQuery, " ")
	selCntErr := db.DB().QueryRow(selectComicCountQuery).Scan(&count)
	if selCntErr != nil {
		if selCntErr == sql.ErrNoRows {
			count = 0
		} else {
			log.Fatal(selCntErr)
		}
	}
	return count, whereClause
}

func ComicPrevImg(comicPath string) string {
	entries, entryErr := os.ReadDir(comicPath)
	if entryErr != nil {
		log.Fatal(entryErr)
	}

	dirs := []string{}
	imgs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
		if strings.Contains(entry.Name(), ".webp") {
			imgs = append(imgs, entry.Name())
		}
	}
	if len(imgs) > 0 {
		comicPrevImg := strings.Join([]string{comicPath, imgs[0]}, "")
		if _, comicImg := os.Stat(comicPrevImg); !os.IsNotExist(comicImg) {
			return comicPrevImg
		}
	}
	if len(imgs) == 0 && len(dirs) > 0 {
		return ComicPrevImg(strings.Join([]string{comicPath, dirs[0], "/"}, ""))
	}
	return ""
}

func ComicContent(w http.ResponseWriter, r *http.Request) string {
	var comic ComicDetails
	var reqComic RequestComicDetails
	json.NewDecoder(r.Body).Decode(&reqComic)

	comicQuery := "SELECT title, tag, translated, author, rating FROM comic WHERE title = ?"
	selErr := db.DB().QueryRow(comicQuery, reqComic.ComicName).Scan(&comic.Title, &comic.Tag, &comic.Translated, &comic.Author, &comic.Rating)
	if selErr != nil {
		if selErr == sql.ErrNoRows {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return "Comic Response Failure"
		} else {
			log.Fatal(selErr)
		}
	}
	comicPath := strings.Join([]string{"/media/comic/", reqComic.ComicName, "/"}, "")
	comic.PrevImgPath = ComicPrevImg(comicPath)

	entries, entryErr := os.ReadDir(comicPath)
	if entryErr != nil {
		log.Fatal(entryErr)
	}
	dirs := []string{}
	imgs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
		if strings.Contains(entry.Name(), ".webp") {
			imgs = append(imgs, entry.Name())
		}
	}

	if len(imgs) > 0 && len(dirs) == 0 {
		var comicPart ComicPart
		var comicPage ComicPage
		comicPart.Subtitle = ""

		for i, img := range imgs {
			comicPage.Id = i
			comicPage.ImgPath = strings.Join([]string{comicPath, img}, "")
			comicPart.Pages = append(comicPart.Pages, comicPage)
		}
		comic.Parts = append(comic.Parts, comicPart)
	}

	if len(imgs) == 0 && len(dirs) > 0 {
		for _, dir := range dirs {
			dirName := strings.Join([]string{comicPath, dir, "/"}, "")
			entries, entryErr := os.ReadDir(dirName)
			if entryErr != nil {
				log.Fatal(entryErr)
			}

			var comicPart ComicPart
			comicPart.Subtitle = dir
			for i, entry := range entries {
				var comicPage ComicPage
				if strings.Contains(entry.Name(), ".webp") {
					comicPage.Id = i
					comicPage.ImgPath = strings.Join([]string{dirName, entry.Name()}, "")
					comicPart.Pages = append(comicPart.Pages, comicPage)
				}
			}
			comic.Parts = append(comic.Parts, comicPart)
		}
	}
	json.NewEncoder(w).Encode(comic)
	return "Comic Response"
}

func ComicUpdate(w http.ResponseWriter, r *http.Request) string {
	var comicData RequestComicUpdate
	json.NewDecoder(r.Body).Decode(&comicData)
	res := Response{1}

	if comicData.Action == "refresh" {
		if ComicListRefresh() {
			res.Status = 0
		}
	}
	if comicData.Action == "edit" {
		if ComicInfoEdit(comicData.Title, comicData.Tag, comicData.Translated, comicData.Author, comicData.Rating) {
			res.Status = 0
		}
	}
	json.NewEncoder(w).Encode(res)
	return "Comic Updated"
}