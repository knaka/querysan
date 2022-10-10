package qsfts

import (
	"fmt"
	"github.com/knaka/querysan/qsfts/models"
	"github.com/rjeczalik/notify"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"log"
	"os"
)

func removeRec(p string) error {
	log.Println("再帰で消せ", p)
	qsfiles := models.Files(
		models.FileWhere.Path.EQ(p),
		qm.Or(
			fmt.Sprintf("%s LIKE ?", models.FileColumns.Path), p+"/%",
		),
	).AllGP(ctx)
	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error 642f44a (%w)", err)
	}
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		_ = tx.Rollback()
	}()
	for _, file := range qsfiles {
		_, err := tx.Exec(`DELETE FROM file_texts WHERE rowid = ?`, file.TextID)
		if err != nil {
			return fmt.Errorf("error 56991e8 (%w)", err)
		}
		file.DeleteP(ctx, tx)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error 6cba943 (%w)", err)
	}
	return nil
}

func WatchFiles() error {
	var err error
	ch := make(chan notify.EventInfo, 10)

	for _, documentDirectory := range conf.DocumentDirectories {
		err := notify.Watch(documentDirectory.Path+"/...", ch, notify.Write, notify.Remove, notify.Rename)
		if err != nil {
			return fmt.Errorf("error f8c3177 (%w)", err)
		}
	}
	defer notify.Stop(ch)
	for {
		ei := <-ch
		path1 := ei.Path()
		event := ei.Event()
		sys1 := ei.Sys()
		log.Println("Got event:", event, path1, sys1)
		if event&notify.Write != 0 {
			// ファイルへの書き込み。単体更新
			err = updateAFile(path1, true)
			if err != nil {
				return fmt.Errorf("error 8ffa416 (%w)", err)
			}
		}
		if event&notify.Remove != 0 {
			// log.Println("ファイルもしくはディレクトリの削除。パス前方一致削除。再帰で通知が来ていたら、配下のファイルはすでに消されていると思われるので、何もしなくて良いこともある")
			// 消えたのがディレクトリかは分からない。もう無いので
			err = removeRec(path1)
			if err != nil {
				return fmt.Errorf("error 9de8011 (%w)", err)
			}
		}
		if event&notify.Rename != 0 {
			log.Println("It's a Rename event")
			stat, err := os.Stat(path1)
			if os.IsNotExist(err) {
				err = removeRec(path1)
				if err != nil {
					return fmt.Errorf("error c3d4017 (%w)", err)
				}
				return nil
			} else if err != nil {
				return fmt.Errorf("error 4b3017f (%w)", err)
			}
			if stat.IsDir() {
				// 要・再帰追加
				log.Println(path1, "is a directory. ディレクトリ再帰更新")
				err = ScanFiles(path1)
				if err != nil {
					return fmt.Errorf("error 6013ae8 (%w)", err)
				}
			} else {
				// 要・単体追加
				log.Println(path1, "is not a directory. ファイル単体更新")
				err = updateAFile(path1, true)
				if err != nil {
					return fmt.Errorf("error 1e936c1 (%w)", err)
				}
			}
		}
	}
}
