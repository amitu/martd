package main

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"flag"
	"time"
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
	c      *Channel
	m, old *Message
}

func Persist(c *Channel, m, old *Message) {
	PersistChan <- &DMessage{c, m, old}
}

func EmptyChannel(c *Channel) {
	PersistChan <- &DMessage{c, nil, nil}
}

func DumpChannels() {
	PersistChan <- &DMessage{nil, nil, nil}
}

func InsertPayload(dm *DMessage) {
	tx, err := PersistDB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	if dm.c == nil {
		rows, err := tx.Query(
			`select
				id, channel, expiry, size, life, one2one, key, payload
			from payloads order by channel desc`,
		)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int64
			var channel string
			var expiry int64
			var size uint
			var life int64
			var one2one bool
			var key string
			var payload []byte
			rows.Scan(
				&id, &channel, &expiry, &size, &life, &one2one, &key, &payload,
			)
			log.Println(
				channel, expiry, size, life, one2one, key, id, string(payload),
			)
		}
		return
	}

	if dm.m == nil {
		stmt, err := tx.Prepare("delete from payloads where channel = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(dm.c.Name)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	stmt, err := tx.Prepare(
		`insert into payloads(
			id, channel, expiry, size, life, one2one, key, payload
		) values (?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		dm.m.Created, dm.c.Name, dm.m.Created + int64(dm.c.Life), dm.c.Size,
		dm.c.Life, dm.c.One2One, dm.c.Key, dm.m.Data,
	)
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
}

func Persister() {
	var err error
	PersistDB, err = GetDB()
	if err != nil {
		log.Panicln("Could not open DB", err)
	}
	for {
		dm := <- PersistChan
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
			id      integer not null primary key,
			channel text,
			expiry  integer,
			size    integer,
			life    integer, -- number of seconds
			one2one integer,
			key     text,
			payload blob
		);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Println("Table already exists:", err)
	} else {
		log.Println("Table created.")
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


	stmt, err := db.Prepare("delete from payloads where expiry < ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now().UnixNano())
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(
		`select
			id, channel, expiry, size, life, one2one, key, payload
		from payloads`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var channel string
		var expiry int64
		var size uint
		var life int64
		var one2one bool
		var key string
		var payload []byte
		rows.Scan(
			&id, &channel, &expiry, &size, &life, &one2one, &key, &payload,
		)
		log.Println(
			channel, expiry, size, life, one2one, key, id, string(payload),
		)
		ch, err := GetOrCreateChannel(
			channel, size, time.Duration(life), one2one, key,
		)
		if err != nil {
			log.Fatalln("Error loading channel:", err)
		}
		log.Println(ch)
		m := &Message{Data: payload, Created: id}
		ch.Messages.Push(m)
	}

	return nil
}
