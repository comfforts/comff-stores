package test

import (
	geo_v1 "github.com/comfforts/comff-geo/api/geo/v1"
)

func BuildPetalumaSet1() []*geo_v1.GeoRequest {
	return []*geo_v1.GeoRequest{
		{
			Street:     "2 Turquoise Ct",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.22507858276367,
			Longitude:  -122.61660766601562,
		},
		{
			Street:     "212 2nd St",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.23274230957031,
			Longitude:  -122.63594055175781,
		},
		{
			Street:     "21 4th St",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.2329613,
			Longitude:  -122.6399594,
		},
		{
			Street:     "931 Petaluma Blvd S",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.2284562,
			Longitude:  -122.6258162,
		},
		{
			Street:     "201 Fair St",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.227476,
			Longitude:  -122.6461669,
			// "dacdbddabcadcddabb"
		},
		{
			Street:     "1160 Schuman Ln",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.2421712,
			Longitude:  -122.657061,
		},
	}
}

func BuildPetalumaSet2() []*geo_v1.GeoRequest {
	return []*geo_v1.GeoRequest{
		{
			Street:     "1280 N McDowell Blvd",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94954",
			Country:    "USA",
			Latitude:   38.2729086,
			Longitude:  -122.6621815,
		},
		{
			Street:     "1371 N McDowell Blvd",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94954",
			Country:    "USA",
			Latitude:   38.2739159,
			Longitude:  -122.6669018,
		},
		{
			Street:     "4995 Petaluma Blvd N",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94952",
			Country:    "USA",
			Latitude:   38.2693353,
			Longitude:  -122.6709554,
		},
		{
			Street:     "1390 N McDowell Blvd STE A",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94954",
			Country:    "USA",
			Latitude:   38.2753438,
			Longitude:  -122.6673901,
		},
		{
			Street:     "50 Ely Rd N",
			City:       "Petaluma",
			State:      "CA",
			PostalCode: "94954",
			Country:    "USA",
			Latitude:   38.2821292,
			Longitude:  -122.6655235,
		},
	}
}
