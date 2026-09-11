package model

// BidRequest represents the OpenRTB 2.5 request payload from an Exchange/SSP.
type BidRequest struct {
	ID     string       `json:"id"`
	Imp    []Impression `json:"imp"`
	Site   *Site        `json:"site,omitempty"`
	Device *Device      `json:"device,omitempty"`
	User   *User        `json:"user,omitempty"`
	At     int          `json:"at,omitempty"`   // Auction type: 1 = First Price, 2 = Second Price
	TMax   int          `json:"tmax,omitempty"` // Timeout in ms
}

type Impression struct {
	ID       string  `json:"id"`
	Banner   *Banner `json:"banner,omitempty"`
	BidFloor float64 `json:"bidfloor,omitempty"`
	Instl    int     `json:"instl,omitempty"` // Interstitial flag
}

type Banner struct {
	W   int `json:"w"`
	H   int `json:"h"`
	Pos int `json:"pos,omitempty"`
}

type Site struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain"`
}

type Device struct {
	UA string `json:"ua,omitempty"`
	IP string `json:"ip,omitempty"`
}

type User struct {
	ID string `json:"id"`
}

// BidResponse represents the OpenRTB 2.5 response returned to the Exchange.
type BidResponse struct {
	ID      string    `json:"id"`
	SeatBid []SeatBid `json:"seatbid"`
	BidID   string    `json:"bidid,omitempty"`
	Cur     string    `json:"cur,omitempty"` // Currency, e.g., "USD"
	NBR     int       `json:"nbr,omitempty"` // No-Bid Reason code
}

type SeatBid struct {
	Bid  []Bid  `json:"bid"`
	Seat string `json:"seat,omitempty"`
}

type Bid struct {
	ID    string   `json:"id"`
	ImpID string   `json:"impid"`
	Price float64  `json:"price"`
	AdID  string   `json:"adid,omitempty"`
	NURL  string   `json:"nurl,omitempty"`
	ADM   string   `json:"adm,omitempty"`
	IURL  string   `json:"iurl,omitempty"`
	CID   string   `json:"cid,omitempty"`
	CRID  string   `json:"crid,omitempty"`
	Cat   []string `json:"cat,omitempty"`
	W     int      `json:"w,omitempty"`
	H     int      `json:"h,omitempty"`
}
