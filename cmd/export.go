package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"golang.org/x/term"
)

type DatabaseColumn struct {
	column_name    string
	column_default string
	is_nullable    string
	data_type      string
	column_type    string
	column_key     string
}

type DatabaseTable struct {
	name    string
	columns []DatabaseColumn
}

const outputFilePermissions = 0644
const maxSheetNameLength = 31

// Use var instead of const for SQL queries because heredoc.Doc returns
// a string that cannot be evaluated at compile-time
var tableNamesSql = heredoc.Doc(`
	SELECT
		table_name
	FROM
		information_schema.tables
	WHERE
		table_schema = ?
	ORDER BY
		table_name ASC
`)
var tableColumnsSql = heredoc.Doc(`
	SELECT
		column_name,
		COALESCE(column_default, 'NULL'),
		is_nullable,
		data_type,
		column_type,
		column_key
	FROM
		information_schema.columns
	WHERE
		table_schema = ?
		AND table_name = ?
`)

var username string
var database string
var hostname string
var port int
var outputFile string
var passwordPrompt bool
var tableIncludes []string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database structure to Excel",
	Run: func(cmd *cobra.Command, args []string) {
		tables := []DatabaseTable{}
		password := ""

		if passwordPrompt {
			fmt.Print("Enter password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			password = string(passwordBytes)

			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println()

		dsn := username + ":" + password + "@tcp(" + hostname + ":" + strconv.Itoa(port) + ")/" + database
		db, err := sql.Open("mysql", dsn)

		if err != nil {
			log.Println(dsn)
			log.Fatal(err)
		}

		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)

		tableNamesRows, err := db.Query(tableNamesSql, database)
		if err != nil {
			log.Fatal(err)
		}

		for tableNamesRows.Next() {
			table := DatabaseTable{}

			err := tableNamesRows.Scan(&table.name)

			if err != nil {
				log.Fatal(err)
			}

			tables = append(tables, table)
		}

		for t := range tables {
			if inTableIncludes(tableIncludes, tables[t].name) {
				columnRows, err := db.Query(tableColumnsSql, database, tables[t].name)

				if err != nil {
					log.Fatal(err)
				}

				for columnRows.Next() {
					column := DatabaseColumn{}

					err := columnRows.Scan(
						&column.column_name,
						&column.column_default,
						&column.is_nullable,
						&column.data_type,
						&column.column_type,
						&column.column_key,
					)

					if err != nil {
						log.Fatal(err)
					}

					tables[t].columns = append(tables[t].columns, column)
				}
			}
		}

		// We have the structures for all the tables, so close the database
		// connection and start building the XLSX file
		db.Close()

		f := excelize.NewFile()

		boldStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		for t := range tables {
			if len(tables[t].columns) >= 1 {
				sheetName := tables[t].name
				if len(sheetName) > maxSheetNameLength {
					sheetName = tables[t].name[:maxSheetNameLength]
				}

				_, err := f.NewSheet(sheetName)

				if err != nil {
					log.Fatal(err)
				}

				f.SetCellValue(sheetName, "A1", "Column Name")
				f.SetCellValue(sheetName, "B1", "Column Default")
				f.SetCellValue(sheetName, "C1", "Is Nullable")
				f.SetCellValue(sheetName, "D1", "Data Type")
				f.SetCellValue(sheetName, "E1", "Column Type")
				f.SetCellValue(sheetName, "F1", "Column Key")

				f.SetCellStyle(sheetName, "A1", "F1", boldStyle)

				for c := range tables[t].columns {
					row := strconv.Itoa(c + 2)

					f.SetCellValue(sheetName, "A"+row, tables[t].columns[c].column_name)
					f.SetCellValue(sheetName, "B"+row, tables[t].columns[c].column_default)
					f.SetCellValue(sheetName, "C"+row, tables[t].columns[c].is_nullable)
					f.SetCellValue(sheetName, "D"+row, tables[t].columns[c].data_type)
					f.SetCellValue(sheetName, "E"+row, tables[t].columns[c].column_type)
					f.SetCellValue(sheetName, "F"+row, tables[t].columns[c].column_key)
				}
			}
		}

		// Delete the first sheet if it is the default
		defaultSheet := f.GetSheetName(0)
		if defaultSheet == "Sheet1" {
			err := f.DeleteSheet(defaultSheet)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err := f.SaveAs(outputFile); err != nil {
			log.Fatal(err)
		}

		if err := os.Chmod(outputFile, outputFilePermissions); err != nil {
			log.Fatal(err)
		}

		f.Close()
	},
}

func init() {
	exportCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	exportCmd.MarkFlagRequired("username")

	exportCmd.Flags().StringVarP(&database, "database", "D", "", "Database name")
	exportCmd.MarkFlagRequired("database")

	exportCmd.Flags().StringVar(&hostname, "host", "localhost", "Hostname")

	exportCmd.Flags().IntVarP(&port, "port", "P", 3306, "Port number")

	exportCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file")
	exportCmd.MarkFlagRequired("output-file")

	exportCmd.Flags().BoolVarP(&passwordPrompt, "password", "p", false, "Prompt for password")

	exportCmd.Flags().StringSliceVar(&tableIncludes, "table", []string{}, "Only include these tables")

	rootCmd.AddCommand(exportCmd)
}

func inTableIncludes(tableIncludes []string, tableName string) bool {
	if len(tableIncludes) == 0 {
		return true
	} else {
		return slices.Contains(tableIncludes, tableName)
	}
}
