package main

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"flag"
)

var (
	PersistChan chan *DMessage
	PersistFile string
	PersistDB   *sql.DB
)

func init() {
	flag.StringVar(&PersistFile, "persist", "persist.db", "Persist File")
	PersistChan = make(chan *DMessage)
}

type DMessage struct {
	channel string
	m, old  *Message
}

func Persist(channel string, m, old *Message) {
	PersistChan <- &DMessage{channel, m, old}
}

func InsertPayload(dm *DMessage) {
	tx, err := PersistDB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(
		"insert into payloads(id, channel, payload) values(?, ?, ?)",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(dm.m.Created, dm.channel, dm.m.Data)
	if err != nil {
		log.Fatal(err)
	}

	if dm.old != nil {
		stmt, err := tx.Prepare("delete from payloads where id = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(dm.old.Created)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}

func Persister() {
	var err error
	PersistDB, err = GetDB()
	if err != nil {
		log.Panicln("Could not open DB", err)
	}
	for {
		dm := <- PersistChan
		log.Println(dm)
		InsertPayload(dm)
	}
}

func GetDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", PersistFile)
	if err != nil {
		log.Println("Failed to open DB", err)
		return db, err
	}

	// check if table exists, if not create it
	sqlStmt := `
		create table payloads (
			id integer not null primary key,
			channel text,
			payload blob
		);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("Table already exists.")
	} else {
		log.Printf("Table created.")
	}

	return db, nil
}

func ReadChannels() error {
	db, err := GetDB()
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()

	rows, err := db.Query("select id, channel, payload from payloads")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var payload []byte
		var channel string
		rows.Scan(&id, &channel, &payload)
		log.Println(channel, id, string(payload))
	}

	return nil
}
