package types

// ExpandRequest represents an expand request struct in the server library backend.
type ExpandRequest [][3]string

func (tuple ExpandRequest) Len() int {
	return len(tuple)
}

func (tuple ExpandRequest) Less(i, j int) bool {
	return tuple[i][0]+tuple[i][1]+tuple[i][2] < tuple[j][0]+tuple[j][1]+tuple[j][2]
}

func (tuple ExpandRequest) Swap(i, j int) {
	tuple[i], tuple[j] = tuple[j], tuple[i]
}
