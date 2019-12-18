package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

type File struct {
	// File Id in the storage
	FileId string
	// Owner's id of the file
	Owner uint
	// Original name of the file
	Filename string
	// Size of the file
	Size int64
	// Is the file's upload finished
	Completed bool
}

func CreateFilesTable() {
	const query = "CREATE TABLE IF NOT EXISTS files ( id serial PRIMARY KEY, file_id text NOT NULL UNIQUE, owner serial NOT NULL, filename text NOT NULL, size bigint, completed bool )"
	if _, err := DB.Exec(query); err != nil {
		panic(err)
	}
}

func (f *File) Create() error {
	const query = "INSERT INTO files (file_id, owner, filename, size, completed) VALUES ($1, $2, $3, $4, $5)"
	if _, err := DB.Exec(query, f.FileId, f.Owner, f.Filename, f.Size, f.Completed); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create file %s (with id %s) for user %d\n", f.Filename, f.FileId, f.Owner)
		return err
	}
	return nil
}

func DeleteFile(owner uint, fileid string) error {
	var db_owner uint
	const validate = "SELECT owner FROM files WHERE file_id=$1"
	row := DB.QueryRow(validate, fileid)
	err := row.Scan(&db_owner)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "Could not get file %s (claiming owner %d)\n", fileid, owner)
			return err
		} else {
			panic(err)
		}
	}
	if db_owner != owner {
		return errors.New("Unauthorized")
	}

	const query = "DELETE FROM files WHERE file_id = $1"
	if _, err := DB.Exec(query, fileid); err != nil {
		fmt.Fprintf(os.Stderr, "Could not delete file with id %s\n", fileid)
		return err
	}
	return nil
}

func GetFiles(user uint) ([]File, error) {
	files := make([]File, 0)
	rows, err := DB.Query("SELECT file_id, owner, filename, size, completed FROM files WHERE owner = $1", user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get files for user %d\n", user)
		return files, err
	}
	defer rows.Close()
	for rows.Next() {
		var fileid string
		var owner uint
		var size int64
		var completed bool
		err = rows.Scan(&fileid, &owner, &size, &completed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file for user %d\n", user)
		} else {
			file := File{FileId: fileid, Owner: owner, Size: size, Completed: completed}
			files = append(files, file)
		}
	}
	return files, nil
}

func SetFileCompleted(fileid string) error {
	const query = "UPDATE files SET completed=true WHERE file_id = $1"
	if _, err := DB.Exec(query, fileid); err != nil {
		fmt.Fprintf(os.Stderr, "Could not set as completed file %s", fileid)
		return err
	}
	return nil
}
