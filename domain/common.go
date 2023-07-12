package domain

type Response struct {
	Code     int
	GRPCCode int
	Message  string
	Data     interface{}
}

type VerboseGreetingRequest struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	FavoriteGame Game   `json:"favorite_game"`
}

type Game struct {
	Name    string `json:"name"`
	Console string `json:"console"`
}

type Villager struct {
	Name        string `json:"name"`
	Personality string `json:"personality"`
}

type FindStreamClientSideRequest struct {
	Name []string `json:"name"`
}
