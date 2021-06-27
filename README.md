# State
State is a package for managing state.

## Cells
Cells are compare-and-swap cells.  They support reading and compare-and-swapping the data.

## Content-Addressed Data
Content-addressed data stores identify pieces of data by their hash.
Posting to the store returns the hash of the data, which can be used later to retrieve the data, and verify that it is correct.
