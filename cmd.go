package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global Parameter
var (
	cmdOptions Command
)

// Root command line options
type Command struct {
	Debug            bool
	Username         string
	Password         string
	Database         string
	Hostname         string
	Port             int
	DB               Database
	Tab              Tables
	Rows             int
	IgnoreConstraint bool
	DontPrompt       bool
}

// Database command line options
type Database struct {
	FakeDB          bool
	FakeDBTableRows bool
}

// Table command line options
type Tables struct {
	FakeNewTables    bool
	TotalTables      int
	MaxColumns       int
	CaseSensitive    bool
	TableNamePrefix  string
	ColumnNamePrefix string
	SchemaName       string
	FakeTablesRows   string
}

// The root commands.
var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s", programName),
	Short: "Insert random data into the postgres database",
	Long: "This program generates fake data into a postgres database cluster. \n" +
		"PLEASE DO NOT run on a mission critical databases",
	Version: programVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Before running any command setup the logger log level
		initLogger(cmdOptions.Debug)

		// if the rows are set to below 1, then error out
		if cmdOptions.Rows < 1 {
			Fatalf("Argument Error: minimum row cannot be less than 1")
		}

		// Ensure we can make a successful connection to the database
		// by printing the version of the database we are going to mock
		dbVersion()
	},
	Run: func(cmd *cobra.Command, args []string) {
		Fatalf("No sub commands used, please run \"%s --help\" for all the options", programName)
	},
}

// The database sub commands
var databaseCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{`d`},
	Short:   "Mock at database level",
	Long:    "Creates a fake tables mimicking a real life database & also can mock the whole database",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Error out if no flags set
		if !cmdOptions.DB.FakeDBTableRows && !cmdOptions.DB.FakeDB {
			Fatalf("No flags set, run \"%s database --help\" for all options for this sub command", programName)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// if there is a request to create a fake database, run the demo database script
		if cmdOptions.DB.FakeDB {
			ExecuteDemoDatabase()
		}
		// Mock the full database
		if cmdOptions.DB.FakeDBTableRows {
			MockDatabase()
		}
	},
}

// The tables sub commands
var tablesCmd = &cobra.Command{
	Use:     "tables",
	Aliases: []string{`t`},
	Short:   "Mock at table level",
	Long:    "Creates fake tables, or mock tables with fake data",
	PreRun: func(cmd *cobra.Command, args []string) {
		// no parameter is given
		if !cmdOptions.Tab.FakeNewTables && IsStringEmpty(cmdOptions.Tab.FakeTablesRows) {
			Fatalf("No flags set, run \"%s tables --help\" for all options for this sub command", programName)
		}
		// either create fake tables or insert mock table rows are allowed, not together
		if cmdOptions.Tab.FakeNewTables && !IsStringEmpty(cmdOptions.Tab.FakeTablesRows) {
			Fatalf("Cannot perform create table & mock tables together, choose one", programName)
		}
		// if there is request for new tables and no of tables parameter is below 1 then error out
		if cmdOptions.Tab.FakeNewTables && cmdOptions.Tab.TotalTables < 1 {
			Fatalf("Cannot have total number of tables below 1, please check the arguments")
		}
		// if there is request for new tables and no of tables parameter is below 1 then error out
		if cmdOptions.Tab.FakeNewTables && cmdOptions.Tab.MaxColumns < 10 {
			Fatalf("Cannot have max number of columns below 10, please check the arguments")
		}
		// If the mock table name or columns are empty then error out
		if IsStringEmpty(cmdOptions.Tab.ColumnNamePrefix) || IsStringEmpty(cmdOptions.Tab.TableNamePrefix) {
			Fatalf("Cannot have the column or table prefix empty, please check the arguments")
		}
		// If schema name is empty
		if IsStringEmpty(cmdOptions.Tab.SchemaName) {
			Fatalf("Cannot have the schema name empty, please check the arguments")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If there is a request to create fake tables then
		if cmdOptions.Tab.FakeNewTables {
			CreateFakeTables()
		}
		// if there is a request to mock few tables
		if !IsStringEmpty(cmdOptions.Tab.FakeTablesRows) {
			MockTables()
		}
	},
}

// Initialize the cobra command line
func init() {
	// Load the environment variable using viper
	viper.AutomaticEnv()

	// Root command flags
	rootCmd.PersistentFlags().BoolVarP(&cmdOptions.Debug, "verbose", "v",
		false, "Enable verbose or debug logging")
	rootCmd.PersistentFlags().IntVarP(&cmdOptions.Rows, "rows", "r",
		10, "Total rows to be faked or mocked")
	rootCmd.PersistentFlags().IntVarP(&cmdOptions.Port, "port", "p",
		viper.GetInt("PGPORT"), "Port number of the postgres database")
	rootCmd.PersistentFlags().StringVarP(&cmdOptions.Hostname, "address", "a",
		viper.GetString("PGHOSTADDR"), "Hostname where the postgres database lives")
	rootCmd.PersistentFlags().StringVarP(&cmdOptions.Username, "username", "u",
		viper.GetString("PGUSER"), "Username to connect to the database")
	rootCmd.PersistentFlags().StringVarP(&cmdOptions.Password, "password", "w",
		viper.GetString("PGPASSWORD"), "Password for the user to connect to database")
	rootCmd.PersistentFlags().StringVarP(&cmdOptions.Database, "database", "d",
		viper.GetString("PGDATABASE"), fmt.Sprintf("Database to %s the data", programName))
	rootCmd.PersistentFlags().BoolVarP(&cmdOptions.IgnoreConstraint, "ignore", "i",
		false, "Ignore checking and fixing constraints")
	rootCmd.PersistentFlags().BoolVarP(&cmdOptions.DontPrompt, "dont-prompt", "q",
		false, "Run without asking for confirmation")

	// Attach the sub commands
	rootCmd.AddCommand(databaseCmd)
	rootCmd.AddCommand(tablesCmd)

	// Database command flags
	databaseCmd.Flags().BoolVarP(&cmdOptions.DB.FakeDB, "create-db", "c", false,
		"Create fake tables mimicking a real life database")
	databaseCmd.Flags().BoolVarP(&cmdOptions.DB.FakeDBTableRows, "full-database", "f", false,
		"Fake all the tables in the database with fake data")

	// Table command flags
	tablesCmd.Flags().BoolVarP(&cmdOptions.Tab.FakeNewTables, "create-tables", "c", false,
		"Create fake tables in the database")
	tablesCmd.Flags().IntVarP(&cmdOptions.Tab.TotalTables, "num-tables", "n", 10,
		"How many fake tables is needed?")
	tablesCmd.Flags().IntVarP(&cmdOptions.Tab.MaxColumns, "max-table-columns", "m", 10,
		"Max number of columns that is needed i.e columns can be from 1 upto this max value")
	tablesCmd.Flags().BoolVarP(&cmdOptions.Tab.CaseSensitive, "case-sensitive-table-name", "j",
		false, "Table name with only lowercase or a mix of lower and uppercase")
	tablesCmd.Flags().StringVarP(&cmdOptions.Tab.TableNamePrefix, "table-name-prefix", "x",
		"mock_data", "Prefix the mocked table with this name")
	tablesCmd.Flags().StringVarP(&cmdOptions.Tab.ColumnNamePrefix, "column-name-prefix", "y",
		"mock_data", "Prefix the mocked table columns with this name")
	tablesCmd.Flags().StringVarP(&cmdOptions.Tab.SchemaName, "schema-name", "s",
		"public", "Under which schema do these fake tables need to be created or mocked?")
	tablesCmd.Flags().StringVarP(&cmdOptions.Tab.FakeTablesRows, "mock-tables", "t", "",
		"Fake selected list of tables with fake data, to add in multiple tables use \",\" b/w table names ")
}