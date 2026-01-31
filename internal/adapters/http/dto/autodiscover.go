package dto

import "encoding/xml"

// Mozilla Autoconfig

type ClientConfig struct {
	XMLName       xml.Name      `xml:"clientConfig"`
	Version       string        `xml:"version,attr"`
	EmailProvider EmailProvider `xml:"emailProvider"`
}

type EmailProvider struct {
	ID             string         `xml:"id,attr"`
	Domain         string         `xml:"domain"`
	DisplayName    string         `xml:"displayName"`
	IncomingServer IncomingServer `xml:"incomingServer"`
	OutgoingServer OutgoingServer `xml:"outgoingServer"`
}

type IncomingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

type OutgoingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

// Microsoft Autodiscover Request

type AutodiscoverRequest struct {
	XMLName xml.Name `xml:"Autodiscover"`
	Request Request  `xml:"Request"`
}

type Request struct {
	EMailAddress             string `xml:"EMailAddress"`
	AcceptableResponseSchema string `xml:"AcceptableResponseSchema"`
}

// Microsoft Autodiscover Response

type AutodiscoverResponse struct {
	XMLName  xml.Name `xml:"http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006 Autodiscover"`
	Response Response `xml:"http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a Response"`
}

type Response struct {
	Account Account `xml:"Account"`
}

type Account struct {
	AccountType string     `xml:"AccountType"`
	Action      string     `xml:"Action"`
	Protocol    []Protocol `xml:"Protocol"`
}

type Protocol struct {
	Type         string `xml:"Type"`
	Server       string `xml:"Server"`
	Port         int    `xml:"Port"`
	SSL          string `xml:"SSL"`
	AuthRequired string `xml:"AuthRequired"`
}
