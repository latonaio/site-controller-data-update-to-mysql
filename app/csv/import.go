package csv

import (
	"fmt"
	"os"

	"golang.org/x/xerrors"

	"github.com/gocarina/gocsv"
)

type ReservationData struct {
	Notice string `csv:"xxxxxx"`

	sampleCode               string `csv:"sample"`
	`...`
}

func ImportFromLincoln(csvPath string) ([]*ReservationData, error) {
	file, err := os.OpenFile(csvPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, xerrors.New("cannot open csv file")
	}
	defer file.Close()

	reservations := []*ReservationData{}

	if err := gocsv.UnmarshalFile(file, &reservations); err != nil { 
		return nil, fmt.Errorf("cannot decode csv file to struct: %v\n", err)
	}

	return reservations, nil
}
