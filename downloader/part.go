package downloader

type Part struct {
	Index      int
	Start      int64
	End        int64
	ReadLength int64
	Filename   string
}

func (p *Part) size() int64 {
	return p.End - p.Start + 1
}

func (p *Part) isCompleted() bool {
	return p.ReadLength == p.size()
}
