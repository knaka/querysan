package qsfts

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/knaka/querysan/qsfts/models"
	"github.com/samber/lo"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func removeNotFoundEntries() error {
	files := models.Files().AllP(ctx, dbConn)
	// todo: tx？
	for _, file := range files {
		_, err := os.Stat(file.Path)
		if os.IsNotExist(err) {
			log.Println("Removing:", file.Path)
			_, err := dbConn.Exec(`DELETE FROM file_texts WHERE rowid = ?`, file.TextID)
			if err != nil {
				return fmt.Errorf("error 3a6e0a9 (%w)", err)
			}
			file.DeleteP(ctx, dbConn)
		} else if err != nil {
			return fmt.Errorf("error dce7500 (%w)", err)
		}
	}
	return nil
}

func updateAFile(filePath string, force bool) error {
	var err error
	var extensions []string = nil
	for _, documentDirectory := range conf.DocumentDirectories {
		// todo: どうも Go の path は string なので、正規化や / の有無がずっと付いてまわるの嫌だな…何とかならんのか
		if strings.Index(filePath, documentDirectory.Path) == 0 {
			extensions = documentDirectory.Extensions
		}
	}
	if extensions == nil {
		return nil
	}
	ext := path.Ext(filePath)
	// 拡張子が対象外でも early return
	if !lo.Contains(extensions, ext) {
		return nil
	}
	var qsfile *models.File
	qsfile, err = models.FindFileG(ctx, filePath)
	if errors.Is(err, sql.ErrNoRows) {
		qsfile = &models.File{
			// “If the table is initially empty, then a ROWID of 1 is used” ですって
			TextID: 0,
			Path:   filePath,
		}
		// 新規追加開始
		log.Println("Caching:", filePath)
	} else if err != nil {
		return fmt.Errorf("error cd23abc (%w)", err)
	} else {
		if !force {
			// 更新日時を確認し、更新されていなければ early return
			fileInfo, err := os.Stat(filePath)
			// 追加直後に消えている可能性はある
			if os.IsNotExist(err) {
				return nil
			} else if err != nil {
				return fmt.Errorf("error 291d978 (%w)", err)
			}
			modTime := fileInfo.ModTime()
			if modTime.Before(qsfile.UpdatedAt) {
				return nil
			}
		}
		// 更新開始
		log.Println("Updating:", filePath)
	}
	bytes, err := os.ReadFile(filePath)
	log.Println("Size:", len(bytes))
	if err != nil {
		return fmt.Errorf("error 345c3a4 (%w)", err)
	}
	text := divideJapaneseToWordsWithZwsp(string(bytes))
	// todo タイトル抽出
	title := ""
	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error c57cbde (%w)", err)
	}
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		_ = tx.Rollback()
	}()
	var res sql.Result
	if qsfile.TextID == 0 {
		res, err = tx.Exec(`INSERT INTO file_texts VALUES(?, ?)`, title, text)
	} else {
		res, err = tx.Exec(`UPDATE file_texts SET title =?, body = ? WHERE rowid = ?`, title, text, qsfile.TextID)
	}
	if err != nil {
		return fmt.Errorf("error 3c25d4c (%w)", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error bb04c69 (%w)", err)
	}
	if rowsAffected != 1 {
		panic("panic 7909309")
	}
	if qsfile.TextID == 0 {
		textId, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("error c1a95da (%w)", err)
		}
		qsfile.TextID = textId
		qsfile.InsertP(ctx, tx, boil.Infer())
	} else {
		qsfile.UpdateP(ctx, tx, boil.Infer())
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error 124da73 (%w)", err)
	}
	return nil
}

func walkToIndexFile(pathArg string, fileInfo fs.FileInfo, _ error) error {
	if fileInfo.IsDir() {
		return nil
	}
	return updateAFile(pathArg, false)
}

func ScanFiles(dir string) error {
	var err error
	err = removeNotFoundEntries()
	if err != nil {
		return fmt.Errorf("error e6de913 (%w)", err)
	}
	err = filepath.Walk(dir, walkToIndexFile)
	if err != nil {
		return fmt.Errorf("error 01326c0 (%w)", err)
	}
	return nil
}

type QueryResult struct {
	Path    string `boil:"path" json:"path"`
	Title   string `boil:"title" json:"title"`
	Offsets string `boil:"offsets" json:"offsets"`
	Snippet string `boil:"snippet" json:"snippet"`
}

func Query(query string) []*QueryResult {
	query = strings.TrimSpace(query)
	queryDivided := divideJapaneseToWords(query)
	var resultSlice []*QueryResult
	// noinspection SqlResolve
	queries.Raw(`
SELECT path, title, offsets(file_texts) AS offsets, snippet(file_texts) AS snippet
FROM file_texts INNER JOIN files ON file_texts.docid = files.text_id
WHERE file_texts MATCH ?
ORDER BY path`, queryDivided).BindP(ctx, dbConn, &resultSlice)
	if len(resultSlice) == 0 {
		return nil
	}
	return resultSlice
}
