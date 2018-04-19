package jsonapi

type Controller struct {
	APIURLBase string
	Models     ModelMap
}

func (c *Controller) SetAPIURL(url string) error {

}
