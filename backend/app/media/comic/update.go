package comic

import (
	"archive/zip"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
	"os"
	"os/exec"
	"log"
	"path/filepath"
	"database/sql"
	"example.com/m/db"
	"github.com/kjk/lzmadec"
	"github.com/chai2010/webp"
	"golang.org/x/text/encoding/korean"
    "golang.org/x/text/transform"
)

func AnsiToUni(src string) string {
	got, _, _ := transform.String(korean.EUCKR.NewDecoder(), src)
	return got
}

func DecompressZip(file string) string {
	path := "/media/comic/"
	dir := path + file[:len(file) - len(filepath.Ext(file))] + "/"

	archive, openErr := zip.OpenReader(path + file)
	if openErr != nil {
		log.Fatal(openErr)
	} // compress file open

	if _, existErr := os.Stat(dir); os.IsNotExist(existErr) {
		mdErr := os.Mkdir(dir, 0775)
		if mdErr != nil {
			log.Fatal(mdErr)
		}
	} // create dir

	for _, f := range archive.File {
		filePath := filepath.Join(dir, AnsiToUni(f.Name))

        if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
            return ""
        }
        if f.FileInfo().IsDir() {
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            log.Fatal(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            log.Fatal(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            log.Fatal(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            log.Fatal(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
	return dir
}

func Decompress7z(file string) string {
	path := "/media/comic/"
	dir := path + file[:len(file) - len(filepath.Ext(file))] + "/"

	archive, openErr := lzmadec.NewArchive(path + file)
	if openErr != nil {
		log.Fatal(openErr)
	} // compress file open
	if _, existErr := os.Stat(dir); os.IsNotExist(existErr) {
		mdErr := os.Mkdir(dir, 0775)
		if mdErr != nil {
			log.Fatal(mdErr)
		}
	} // create dir

	for _, entry := range archive.Entries {
		if strings.Count(entry.Path, "/") > 0 {
			extDir := dir + strings.Split(entry.Path, "/")[strings.Count(entry.Path, "/") - 1]
			if _, extExistErr := os.Stat(extDir); os.IsNotExist(extExistErr) {
				extDirErr := os.MkdirAll(extDir, 0775)
				if extDirErr != nil {
					log.Fatal(extDirErr)
				}
			}
		} // create dir for the dir in compressed file
		if _, extDirPass := os.Stat(dir + entry.Path); !os.IsNotExist(extDirPass) {
			continue
		} // skip extract dir in compressed file
		extErr := archive.ExtractToFile(dir + entry.Path, entry.Path)
		if extErr != nil {
			log.Fatal(extErr)
		} // extract
	}
	return dir
}

func PruneDir(directory string) bool {
	dirs := []string{}
	entries, entryErr := os.ReadDir(directory)
	if entryErr != nil {
		log.Fatal(entryErr)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	if len(entries) == 1 && len(dirs) == 1 {
		entries, entryErr := os.ReadDir(directory + dirs[0])
		if entryErr != nil {
			log.Fatal(entryErr)
		}
		for _, entry := range entries {
			src := directory + dirs[0] + "/" + entry.Name()
			dst := directory + entry.Name()
			mvErr := os.Rename(src, dst)
			if mvErr != nil {
				log.Fatal(mvErr)
			}
		}
		rmErr := os.Remove(directory + dirs[0])
		if rmErr != nil {
			log.Fatal(rmErr)
		}
		return true
	}
	return false
} // if there is only one folder directly inside the compressed file

func ImgConverter(directory string) bool {
	dirs := []string{}
	imgs := []string{}
	entries, entryErr := os.ReadDir(directory)
	if entryErr != nil {
		log.Fatal(entryErr)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
		if strings.Contains(strings.ToLower(entry.Name()), ".jpg") {
			imgs = append(imgs, entry.Name())
		}
		if strings.Contains(strings.ToLower(entry.Name()), ".jpeg") {
			imgs = append(imgs, entry.Name())
		}
		if strings.Contains(strings.ToLower(entry.Name()), ".png") {
			imgs = append(imgs, entry.Name())
		}
	}
	if len(dirs) > 0 {
		for _, dir := range dirs {
			ImgConverter(directory + dir + "/")
		}
	}
	if len(imgs) > 0 {
		for i, img := range imgs {
			source := directory + img
			destination := directory + fmt.Sprintf("%03d", i + 1) + ".webp"
			convertErr := ImgToWebp(source, destination)
			if convertErr != nil {
				log.Fatal(convertErr)
			}
			rmErr := os.Remove(source)
			if rmErr != nil {
				log.Fatal(rmErr)
			}
		}
	} else {
		return false
	}
	return true
}

func ImgToWebp(sourceImg, destImg string) error {
	srcImg, srcErr := os.Open(sourceImg)
	if srcErr != nil {
		return srcErr
	}
	defer srcImg.Close()

	img, format, decErr := image.Decode(srcImg)
	if decErr != nil {
		return decErr
	}

	dstImg, dstErr := os.Create(destImg)
	if dstErr != nil {
		return dstErr
	}
	defer dstImg.Close()

	options := &webp.Options{
		Lossless: false,
		Quality: 80,
	}

	if format == "jpeg" || format == "png" {
		encErr := webp.Encode(dstImg, img, options)
		if encErr != nil {
			return encErr
		}
		return nil
	} else {
		return fmt.Errorf("Unsupported format: %s", format)
	}
}

func UpdateComicDB(comicName string) bool {
	insertComicQuery := "INSERT INTO comic (title, tag, translated, author, rating) VALUES (?, ?, 0, ?, 0)"
	insertPrepare, _ := db.DB().Prepare(insertComicQuery)
	_, insErr := insertPrepare.Exec(comicName, "Update needed", "Update needed")
	if insErr != nil {
		log.Fatal(insErr)
	}
	return true
}

func IntermBackup(file string) bool {
	srcPath := "/media/comic/"
	dstPath := "/media/backup/"

	mvErr := os.Rename(srcPath + file, dstPath + file)
	if mvErr != nil {
		log.Fatal(mvErr)
	}
	return true
}

func FinalBackup() bool {
	return true
}


func ComicListRefresh() bool {
	comicPath := "/media/comic/"

	comicOut, cCmdErr := exec.Command("ls", comicPath).Output()
	if cCmdErr != nil {
		log.Fatal(cCmdErr)
	}

	comics := strings.Split(string(comicOut), "\n")
	var comicFiles []string
	for _, comicName := range comics {
		if strings.Contains(comicName, ".zip") {
			comicFiles = append(comicFiles, comicName)
		}
		if strings.Contains(comicName, ".7z") {
			comicFiles = append(comicFiles, comicName)
		}
	}
	for _, comicFile := range comicFiles {
		var queryResult string
		comicName := comicFile[:len(comicFile) - len(filepath.Ext(comicFile))]

		selectComicQuery := "SELECT title FROM comic WHERE title = ?"
		selErr := db.DB().QueryRow(selectComicQuery, comicName).Scan(&queryResult)
		if selErr != nil {
			if selErr == sql.ErrNoRows {
				var dir string
				if filepath.Ext(comicFile) == ".7z" {
					dir = Decompress7z(comicFile)
				}
				if filepath.Ext(comicFile) == ".zip" {
					dir = DecompressZip(comicFile)
				}
				if len(dir) > 1 {
					fmt.Printf(comicFile + " extracted")
					if PruneDir(dir) {
						fmt.Printf(" cleanup")
					}
					if ImgConverter(dir) {
						fmt.Printf(" toWebp")
					}
					if IntermBackup(comicFile) {
						fmt.Printf(" backup")
					}
					if UpdateComicDB(comicName) {
						fmt.Printf(" DBupdate")
					}
					fmt.Println()
				}
			} else {
				log.Fatal(selErr)
			}
		} else {
			continue
		}
	}
	return true
}

func ComicInfoEdit(title string, tag string, translated int, author string, rating int) bool {
	var queryResult string
	selectComicQuery := "SELECT title FROM comic WHERE title = ?"
	selErr := db.DB().QueryRow(selectComicQuery, title).Scan(&queryResult)
	if selErr != nil {
		log.Fatal(selErr)
	}
	updateComicQuery := "UPDATE comic SET tag = ?, translated = ?, author = ?, rating = ? WHERE title = ?"
	updatePrepare, _ := db.DB().Prepare(updateComicQuery)
	_, uptErr := updatePrepare.Exec(tag, translated, author, rating, title)
	if uptErr != nil {
		log.Fatal(uptErr)
	}
	return true
}