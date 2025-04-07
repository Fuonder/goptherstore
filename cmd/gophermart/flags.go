package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
sudo -i -u postgres
psql -U postgres

POSTGRE USER && DATABASE:

create user market with encrypted password '12345678';
create database market;
grant all privileges on database market to market;
\c market;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO market;
GRANT ALL ON SCHEMA public TO market;
GRANT CREATE ON SCHEMA public to market;
GRANT USAGE ON SCHEMA public to market;

*/

var (
	version  = "0.0.1"
	progName = "Fuonder's gopherstore"
	source   = "https://github.com/Fuonder/goptherstore"
)

var usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s\nSource code:\t%s\nVersion:\t%s\nUsage of %s:\n",
		progName,
		source,
		version,
		progName)
	flag.PrintDefaults()
}

var (
	ErrNotFullIP   = errors.New("given ip address and port incorrect")
	ErrInvalidIP   = errors.New("incorrect ip address")
	ErrInvalidPort = errors.New("incorrect port number")
)

type netAddress struct {
	ipaddr string
	port   int
}

func (n *netAddress) String() string {
	return fmt.Sprintf("%s:%d", n.ipaddr, n.port)
}
func (n *netAddress) Set(value string) error {
	value = strings.TrimPrefix(value, "http://")
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("%w: \"%s\"", ErrNotFullIP, value)
	}
	n.ipaddr = values[0]
	if n.ipaddr == "" {
		return fmt.Errorf("%w: \"%s\"", ErrInvalidIP, values[0])
	}
	var err error
	n.port, err = strconv.Atoi(values[1])
	if err != nil {
		return fmt.Errorf("%w: \"%s\"", ErrInvalidPort, values[1])
	}
	return nil
}

type Flags struct {
	APIAddress     netAddress
	AccrualAddress netAddress
	DatabaseDSN    string
	LogLevel       string
	Key            string
}

func (f *Flags) String() string {
	return fmt.Sprintf("APIAddress: %s, "+
		"AccrualAddress: %s, "+
		"DatabaseDSN: %s, "+
		"LogLevel: %s"+
		"Key: %s",
		f.APIAddress.String(),
		f.AccrualAddress.String(),
		f.DatabaseDSN,
		f.LogLevel,
		f.Key,
	)
}

var (
	CliOptions = Flags{
		APIAddress: netAddress{
			ipaddr: "localhost",
			port:   1234,
		},
		AccrualAddress: netAddress{
			ipaddr: "localhost",
			port:   8080,
		},
		//DatabaseDSN: "postgres://market:12345678@localhost:5432/market?sslmode=disable",
		DatabaseDSN: "",
		LogLevel:    "info",
	}
)

func parseFlags() error {
	flag.Usage = usage
	flag.Var(&CliOptions.APIAddress, "a", "ip and port of server in format <ip>:<port>")
	flag.Var(&CliOptions.AccrualAddress, "r", "ip and port of accrual service in format <ip>:<port>")
	flag.StringVar(&CliOptions.DatabaseDSN, "d", "postgres://market:12345678@localhost:5432/market?sslmode=disable", "Database DSN")
	flag.StringVar(&CliOptions.LogLevel, "l", "info", "loglevel")
	flag.StringVar(&CliOptions.Key, "k", "TEST123", "encryption key")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		err := CliOptions.APIAddress.Set(envRunAddr)
		if err != nil {
			return err
		}
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		err := CliOptions.AccrualAddress.Set(envAccrualAddr)
		if err != nil {
			return err
		}
	}
	if envDatabaseDSN := os.Getenv("DATABASE_URI"); envDatabaseDSN != "" {
		CliOptions.DatabaseDSN = envDatabaseDSN
	}
	fmt.Println(CliOptions.DatabaseDSN)

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		CliOptions.LogLevel = envLogLevel
	}
	if envSecret := os.Getenv("SECRET"); envSecret != "" {
		CliOptions.Key = envSecret
	}

	return nil
}
