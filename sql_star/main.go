package main

import (
	"database/sql"
	"dev_tools/files"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	DbUrl   string // db connection string e.g. host=%s port=%d user=%s password=%s dbname=%s sslmode=disable
	Dir     string // which dir to look for sql files
	rootCmd = &cobra.Command{
		Use:   "ss",
		Short: "run sql files against a database",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var connectionString string
			if len(DbUrl) < 1 {
				envCon := getConnectionStringFromEnv()
				if len(envCon) < 1 {
					panic("no connection host supplied")
				} else {
					connectionString = envCon
				}
			} else {
				connectionString = DbUrl
			}
			RunSqlStart(connectionString, Dir)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&DbUrl, "db-url", "u", "", "sql connection string")
	rootCmd.PersistentFlags().StringVarP(&Dir, "sql-dir", "d", ".", "what dir to look for sql files")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func RunSqlStart(connString, dir string) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	findSqlFilesInDirToExecute(db, dir)
}

func findSqlFilesInDirToExecute(db *sql.DB, dir string) {
	if !files.IsDir(dir) {
		if isSqlFile(dir) {
			runFile(db, dir)
		} else {
			panic("invalid sql file provided")
		}
	} else {
		findSqlFilesToExecute(db, dir)
	}
}

func findSqlFilesToExecute(db *sql.DB, dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if isSqlFile(path) {
			runFile(db, path)
		}
		return nil
	})
}

func runFile(db *sql.DB, dir string) {
	println(fmt.Sprintf("runing file %s...", dir))
	fileString, err := files.ReadStringFromFile(dir)
	if err != nil {
		panic(err)
	}

	// not using the result set right now, I want to do something with it in the future
	// todo do something with this
	_, err = db.Exec(fileString)
	if err != nil {
		panic(err)
	}
}

func isSqlFile(loc string) bool {
	if len(loc) < 1 {
		return false
	}
	return strings.ToLower(filepath.Ext(loc)) == ".sql"
}

func getConnectionStringFromEnv() string {
	return os.Getenv("SQL_START_CONNECTION_URL")
}
