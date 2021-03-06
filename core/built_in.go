package core

import "os/user"

/* Built-in */

// Built-in data set
type _BuiltInDataSet struct {
	url    string
	path   string
	sep    string
	header bool
	loader func(string, string, bool) DataSet
}

var builtInDataSets = map[string]_BuiltInDataSet{
	"ml-100k": {
		url:    "https://cdn.sine-x.com/datasets/movielens/ml-100k.zip",
		path:   "ml-100k/u.data",
		sep:    "\t",
		header: false,
		loader: LoadDataFromCSV,
	},
	"ml-1m": {
		url:    "https://cdn.sine-x.com/datasets/movielens/ml-1m.zip",
		path:   "ml-1m/ratings.dat",
		sep:    "::",
		header: false,
		loader: LoadDataFromCSV,
	},
	"ml-10m": {
		url:    "https://cdn.sine-x.com/datasets/movielens/ml-10m.zip",
		path:   "ml-10M100K/ratings.dat",
		sep:    "::",
		header: false,
		loader: LoadDataFromCSV,
	},
	"ml-20m": {
		url:    "https://cdn.sine-x.com/datasets/movielens/ml-20m.zip",
		path:   "ml-20m/ratings.csv",
		sep:    ",",
		header: false,
		loader: LoadDataFromCSV,
	},
	"netflix": {
		url:    "https://cdn.sine-x.com/datasets/netflix/netflix-prize-data.zip",
		path:   "netflix/training_set.txt",
		loader: LoadDataFromNetflixStyle,
	},
	"filmtrust": {
		url:    "https://cdn.sine-x.com/datasets/filmtrust/filmtrust.zip",
		path:   "filmtrust/ratings.txt",
		sep:    " ",
		header: false,
		loader: LoadDataFromCSV,
	},
	"epinions": {
		url:    "https://cdn.sine-x.com/datasets/epinions/epinions.zip",
		path:   "epinions/ratings_data.txt",
		sep:    " ",
		header: true,
		loader: LoadDataFromCSV,
	},
}

// The data directories
var (
	downloadDir string
	dataSetDir  string
	TempDir     string
)

func init() {
	usr, _ := user.Current()
	gorseDir := usr.HomeDir + "/.gorse"
	downloadDir = gorseDir + "/download"
	dataSetDir = gorseDir + "/datasets"
	TempDir = gorseDir + "/temp"
}
