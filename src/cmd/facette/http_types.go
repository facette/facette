package main

type httpTypeRecord struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type httpTypeList []httpTypeRecord

func (l httpTypeList) Len() int {
	return len(l)
}

func (l httpTypeList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l httpTypeList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
