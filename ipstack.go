// Copyright © 2019. All rights reserved.
// Author: Alice Qio.
// Contacts: <qioalice@gmail.com>.
// License: https://opensource.org/licenses/MIT
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom
// the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package ipstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// todo: Tests
// todo: Example

// HOW TO USE THIS LIBRARY?
//
// 1. CREATE ENDPOINT
// Create 'Client' object by 'New' function or initialize the default,
// package level 'Client' object by 'Init' function.
// You can pass token directly to the 'New' or 'Init' method, or through
// 'ParamToken' parametrizer.
// You can also specify special behaviour, like always using HTTPS
// (require non-free tariff), or your own golang http.Client object, or
// if you want, decrease network load by specifing only important fields
// that must be returned from ipstack - only fields that you need, or
// if you want full info, specify as many fields as it possible.
// Use 'ParamFields' to do it.
// NOTE! After client has been created, the first request will already
// performed. More than, client creating operation will fall if the first
// request willn't performed successfully.
// What is the first request? Info about your IP, of course.
// If you do not want to perform that request, just pass
// 'ParamDisableFirstMeCall' parameterizer to the 'New' or 'Init' function.
//
// 2. MAKE REQUEST
//
// -- IP (ip string) (r *Response, e error)
// Get info about 'ip'. You can call this method of 'Client' object
// that you created at the first step, or call package level function,
// if you initialize default client.
// -- IPs (ips ...string) (rs []*Response, e error)
// Get info about few 'ips'. Pass it and separate with a comma.
// -- UpdateMe() (e error)
// Fetch the fresh information about your IP.
//
// 2.1. EXTENDED REQUEST, EXTENDED RESPONSE
//
// By default you just calling methods of some 'Client' object -
// created by you or package-level's initialized.
// But what if you want to make a different request from the rest
// that you usually make?
// Or you take great care about memory and do not want to allocate a few bytes
// to the fields of 'Response' object that willn't be used and you want to
// specify only two/three/n fields that you're required and no one bytes more?
// Or may be I forgot to update this library, API has been changed
// and you want to manually decode JSON response from ipstack?
// Just get the 'tRequest' object! Call 'R' method from 'Client'!
// Do what you want and perform request by methods with name that
// you're already know ('IP', 'IPs') or get fresh info about your IP
// by calling 'Me' method.
// But be careful, by calling one of these method, you'll get 'tRawResponse'
// object that will contain error objects of performed request (if it is),
// and JSON raw data. Take it and do what you want with it.
// If you don't want manually check errors (from API by JSON response)
// and/or manually decode it, just call 'CheckError' and 'DecodeIt' methods
// of 'tRawResponse' object.
// But you must manually create object to that decode process will peform
// and pass its reference to the 'DecodeIt' method.
// 'DecodeIt' always returns an error, and you can chain all that query:
//       result := <result_object_allocate_mem>
//       err := client.R().<methods>(...).CheckError().DecodeIt(&result)
// And if any error occurred before 'CheckError' or 'DecodeIt' has been
// called, 'DecodeIt' will return that error.
// Yes it will pass through whole that way.

// Base of API endpoints.
// All Web API requests will be performed to the endpoints started
// these URL with.
const (
	// Default Web API endpoint
	cApiEndpointHTTP string = "http://api.ipstack.com/"
	// SSL Web API endpoint, disabled by default
	cApiEndpointHTTPS string = "https://api.ipstack.com/"
	// security param
	securityParam string = "security=1"
)

// 'Client' is the class that represents golang point to make Web API
// requests to ipstack.
//
// WARNING! Do not create this object directly! This class has important
// private fields that must be initialized only by constructor.
// Use 'New' function to create a new 'Client' object.
//
// After 'Client' object created, (and if you don't disable first check)
// you already have an information about your IP.
type Client struct {
	me              *Response
	baseReq         *tRequest
	skipInitFetchMe bool
}

// 'tClientParam' is the internal auxiliary type that is alias to the
// function only one argument type of 'Client' by pointer is receive.
// It used to represent some parameters for fabric function 'New' -
// the constructor of 'Client' object.
//
// So, there are few functions using which you can get the 'tClientParam'
// objects and then pass it to the 'New' function ('Client' constructor).
type tClientParam func(c *Client)

// 'tRequest' is the internal private type that represents one request
// to the ipstack Web API.
//
// So, object of this class will be created when 'Client' object will being
// initialize and that object will be marked as 'base request object'.
// All special Web API params, like GET request params, golang http.Client
// object, HTTP or HTTPS schema, etc will be stored to 'tRequest' object.
//
// Read the docs for 'R' method of 'Client' class to understand how
// 'tRequest' object works and why it exists.
type tRequest struct {
	token           string
	client          *http.Client
	endpoint        string
	reqArgs         url.Values
	reqArgsBuilt    string
	securityEnabled bool
}

// 'tResponse' is the internal private type that represents some RAW
// response from ipstack to the some Web API request.
//
// So, for what it needs?
// General, only when you do not want to allocate memory to store and
// represent full response object ('Response' object) or if you want
// decode response (JSON) manually.
//
// NOTE! If you using custom request using 'R' method of 'Client' class,
// you will get 'tRequest' object, and methods 'IP', 'IPs', 'Me' of
// 'tRequest' class always return 'tResponse' object.
// It means, that you must take care of error analysing and JSON decoding.
//
// NOTE! It guarantees, that if 'RequestError' isn't nil, 'ResponseError' and
// 'RawData' are. Similar, if 'ResponseError' isn't nil, 'RawData' is.
type tResponse struct {
	RawData []byte
	Error   error
}

// 'Response' represents the golang view of Web API response.
// Fields 'Location', 'Timezone', 'Currency', 'Connection', 'Security'
// might be nil, if you didn't request it earlier using 'Fields' method
// of 'Client' or 'tRequest' classes, or 'ParamFields' parameter of 'Client'
// constructor ('New' function).
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type Response struct {
	IP            string               `json:"ip"`
	Hostname      string               `json:"hostname"`
	Type          string               `json:"type"`
	ContinentCode string               `json:"continent_code"`
	ContinentName string               `json:"continent_name"`
	CountryCode   string               `json:"country_code"`
	CountryName   string               `json:"country_name"`
	RegionCode    string               `json:"region_code"`
	RegionName    string               `json:"region_name"`
	City          string               `json:"city"`
	Zip           string               `json:"zip"`
	Latitide      float32              `json:"latitude"`
	Longitude     float32              `json:"longitude"`
	Location      *tResponseLoc        `json:"location"`
	Timezone      *tResponseTimeZone   `json:"time_zone"`
	Currency      *tResponseCurrency   `json:"currency"`
	Connection    *tResponseConnection `json:"connection"`
	Security      *tResponseSecurity   `json:"security"`
}

// 'tResponseLoc' is the part of Web API response and represents
// the location info about requested IP.
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type tResponseLoc struct {
	GeonameID               int                `json:"geoname_id"`
	Capital                 string             `json:"capital"`
	Languages               []tResponseLocLang `json:"languages"`
	CountryFlagLink         string             `json:"country_flag"`
	CountryFlagEmoji        string             `json:"country_flag_emoji"`
	CountryFlagEmojiUnicode string             `json:"country_flag_emoji_unicode"`
	CallingCode             string             `json:"calling_code"`
	IsEU                    bool               `json:"is_eu"`
}

// 'tResponseLocLang' is the part of Web API response and represents
// the info about languages in the location of requested IP.
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type tResponseLocLang struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native"`
}

// 'tResponseTimeZone' is the part of Web API response and represents
// the info about time zone info in the location of requested IP.
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type tResponseTimeZone struct {
	ID               string    `json:"id"`
	CurrentTime      time.Time `json:"current_time"`
	GMTOffset        int       `json:"gmt_offset"`
	Code             string    `json:"code"`
	IsDaylightSaving bool      `json:"is_daylight_saving"`
}

// 'tResponseCurrency' is the part of Web API response and represents
// the info about main currency in the location of requested IP.
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type tResponseCurrency struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Plural       string `json:"plural"`
	Symbol       string `json:"symbol"`
	SymbolNative string `json:"symbol_native"`
}

// 'tResponseConnection' is the part of Web API response and represents...
// I dunno what is this part represents.
// If anyone read this section comment and knows, please add info about it.
// PR is welcome!
type tResponseConnection struct {
	ASN int    `json:"asn"`
	ISP string `json:"isp"`
}

// 'tResponseSecurity' is the part of Web API response and represents
// some other info about requested IP.
//
// NOTE! If you do not understand what data stored in field,
// read the docs of the consts 'Field...'  (above).
type tResponseSecurity struct {
	IsProxy     bool        `json:"is_proxy"`
	ProxyType   string      `json:"proxy_type"`
	IsCrawler   bool        `json:"is_crawler"`
	CrawlerName string      `json:"crawler_name"`
	CrawlerType string      `json:"crawler_type"`
	IsTOR       bool        `json:"is_tor"`
	ThreatLevel string      `json:"threat_level"`
	ThreatTypes interface{} `json:"threat_types"`
}

// 'tResponseError' is the internal auxiliary type that exists only for
// represents the parts of JSON encoded response from Web API.
//
// Web API error JSON message looks like:
// "{ "success": false, error: { "code": N, "type": "...", "info": "..." } }".
// The 'code', 'type' and 'info' fields are explains in 'tError' class docs.
type tResponseError struct {
	Success bool   `json:"success"`
	Error   tError `json:"error"`
}

// 'tError' represents the Web API error message.
// Be careful, object of that class might be created only if API request
// has been successfully sent and SOME response has been successfully received.
// And if that response contains error message (that consists of these entities)
// the 'tError' object will be created.
//
// Error entities:
// <code> - is the error code of Web API and the the fastest way
// to distinguish one error from another.
// <type> - is the title of error - a short explanation about
// occurred error - the type of error.
// <info> - is the description of error - a long explanation about
// occurred error - what happened and why.
//
// You can get that error object when you calling 'IP', 'IPs', 'UpdateMe'
// methods as the second return argument, or as field 'ResponseError'
// in 'tRawResponse' object, if you're working with RAW response object
// (often it might be only if you're working with RAW request object
// using 'R' method of 'Client' class).
//
// See 'tResponseError' docs for more details.
type tError struct {
	RawCode int    `json:"code"`
	RawType string `json:"type"`
	RawInfo string `json:"info"`
}

// Predefined consts each of that represents some field in JSON response
// from ipstack on the some request.
//
// So, you can read on https://ipstack.com/documentation that
// you can specify what fields should be returned as response.
// Using that consts and 'Fields' method of 'Client' or 'tRequest' classes
// you can do it.
const (
	// Returns the requested IP address.
	FieldIp string = "ip"
	// Returns the hostname the requested IP resolves to,
	// only returned if Hostname Lookup is enabled.
	FieldHostname string = "hostname"
	// Returns the IP address type IPv4 or IPv6.
	FieldType string = "type"
	// Returns the 2-letter continent code associated with the IP.
	FieldContinentCode string = "continent_code"
	// Returns the name of the continent associated with the IP.
	FieldContinentName string = "continent_name"
	// Returns the 2-letter country code associated with the IP.
	FieldCountryCode string = "country_code"
	// Returns the name of the country associated with the IP.
	FieldCountryName string = "country_name"
	// Returns the region code of the region associated with the IP
	// (e.g. CA for California).
	FieldRegionCode string = "region_code"
	// Returns the name of the region associated with the IP.
	FieldRegionName string = "region_name"
	// Returns the name of the city associated with the IP.
	FieldCity string = "city"
	// Returns the ZIP code associated with the IP.
	FieldZip string = "zip"
	// Returns the latitude value associated with the IP.
	FieldLatitude string = "latitude"
	// Returns the longitude value associated with the IP.
	FieldLongitude string = "longitude"
	// Returns multiple location-related objects
	FieldLocation string = "location"
	// Returns the unique geoname identifier
	// in accordance with the Geonames Registry.z
	FieldLocationGeonameId string = "location.geoname_id"
	// Returns the capital city of the country associated with the IP.
	FieldLocationCapital string = "location.capital"
	// Returns an object containing one or multiple
	// sub-objects per language spoken in the country associated with the IP.
	FieldLocationLanguages string = "location.languages"
	// Returns the 2-letter language code for the given language.
	FieldLocationLanguagesCode string = "location.languages.code"
	// Returns the name (in the API request's main language)
	// of the given language. (e.g. Portuguese)
	FieldLocationLanguagesName string = "location.languages.name"
	// Returns the native name of the given language. (e.g. Português)
	FieldLocationLanguagesNative string = "location.languages.native"
	// Returns an HTTP URL leading to an SVG-flag icon for the country
	// associated with the IP.
	FieldLocationCountryFlag string = "location.country_flag"
	// Returns the emoji icon for the flag of the country associated with the IP.
	FieldLocationCountryFlagEmoji string = "location.country_flag_emoji"
	// Returns the unicode value of the emoji icon for the flag of the country
	// associated with the IP. (e.g. U+1F1F5 U+1F1F9 for the Portuguese flag)
	FieldLocationCountryFlagEmojiUnicode string = "location.country_flag_emoji_unicode"
	// Returns the calling/dial code of the country associated with the IP.
	// (e.g. 351) for Portugal.
	FieldLocationCallingCode string = "location.calling_code"
	// Returns true or false depending on whether or not the county
	// associated with the IP is in the European Union.
	FieldLocationIsEu string = "location.is_eu"
	// Returns an object containing timezone-related data.
	FieldTimeZone string = "timezone"
	// Returns the ID of the time zone associated with the IP.
	// (e.g. America/LosAngeles for PST)
	FieldTimeZoneId string = "timezone.id"
	// Returns the current date and time in the location
	// associated with the IP. (e.g. 2018-03-29T22:31:27-07:00)
	FieldTimeZoneCurrentTime string = "timezone.current_time"
	// Returns the GMT offset of the given time zone in seconds.
	// (e.g. -25200 for PST's -7h GMT offset)
	FieldTimeZoneGmtOffset string = "timezone.gmt_offset"
	// Returns the universal code of the given time zone.
	FieldTimeZoneCode string = "timezone.code"
	// Returns true or false depending on whether or not the given time zone
	// is considered daylight saving time.
	FieldTimeZoneIsDaylightSaving string = "timezone.is_daylight_saving"
	// Returns an object containing currency-related data.
	FieldCurrency string = "currency"
	// Returns the 3-letter code of the main currency associated with the IP.
	FieldCurrencyCode string = "currency.code"
	// Returns the name of the given currency.
	FieldCurrencyName string = "currency.mame"
	// Returns the plural name of the given currency.
	FieldCurrencyPlural string = "currency.plural"
	// Returns the symbol letter of the given currency.
	FieldCurrencySymbol string = "currency.symbol"
	// Returns the native symbol letter of the given currency.
	FieldCurrencySymbolNative string = "currency.symbol_native"
	// Returns an object containing connection-related data.
	FieldConnection string = "connection"
	// Returns the Autonomous System Number associated with the IP.
	FieldConnectionAsn string = "connection.asn"
	// Returns the name of the ISP associated with the IP.
	FieldConnectionIsp string = "connection.isp"
	// Returns an object containing security-related data.
	FieldSecurity string = "security"
	// Returns true or false depending on whether or not the given IP
	// is associated with a proxy.
	FieldSecurityIsProxy string = "security.is_proxy"
	// Returns the type of proxy the IP is associated with.
	FieldSecurityProxyType string = "security.proxy_type"
	// Returns true or false depending on whether or not the given IP
	// is associated with a crawler.
	FieldSecurityIsCrawler string = "security.is_crawler"
	// Returns the name of the crawler the IP is associated with.
	FieldSecurityCrawlerName string = "security.crawler_name"
	// Returns the type of crawler the IP is associated with.
	FieldSecurityCrawlerType string = "security.crawler_type"
	// Returns true or false depending on whether or not the given IP
	// is associated with the anonymous Tor system.
	FieldSecurityIsTor string = "security.is_tor"
	// Returns the type of threat level the IP is associated with.
	FieldSecurityThreatLevel string = "security.threat_level"
	// Returns an object containing all threat types associated with the IP.
	FieldSecurityThreatTypes string = "security.threat_types"
)

// The slice of all predefined field name's consts.
var allFields = []string{
	// General
	FieldIp, FieldHostname, FieldType,
	FieldContinentCode, FieldContinentName,
	FieldCountryCode, FieldCountryName,
	FieldRegionCode, FieldRegionName,
	FieldCity, FieldZip, FieldLatitude, FieldLongitude,
	// Location
	FieldLocation,
	FieldLocationGeonameId, FieldLocationCapital,
	// Location languages
	FieldLocationLanguages,
	FieldLocationLanguagesCode, FieldLocationLanguagesName,
	FieldLocationLanguagesNative,
	// Location
	FieldLocationCountryFlag, FieldLocationCountryFlagEmoji,
	FieldLocationCountryFlagEmojiUnicode, FieldLocationCallingCode,
	FieldLocationIsEu,
	// Time zone
	FieldTimeZone, FieldTimeZoneId, FieldTimeZoneCurrentTime,
	FieldTimeZoneGmtOffset, FieldTimeZoneCode, FieldTimeZoneIsDaylightSaving,
	// Currency
	FieldCurrency, FieldCurrencyCode, FieldCurrencyName, FieldCurrencyPlural,
	FieldCurrencySymbol, FieldCurrencySymbolNative,
	// Connection
	FieldConnection, FieldConnectionAsn, FieldConnectionIsp,
	// Security
	FieldSecurity, FieldSecurityIsProxy, FieldSecurityProxyType,
	FieldSecurityIsCrawler, FieldSecurityCrawlerName, FieldSecurityCrawlerType,
	FieldSecurityIsTor, FieldSecurityThreatLevel, FieldSecurityThreatTypes,
}

// Default client.
// It will be initialized directly by calling 'Init' function,
// or when you will call 'New' function first time and that call will be
// successfull, the 'New' function will also tagged the created client
// as default client (will store pointer to the 'DefaultClient').
var DefaultClient *Client

// 'R' is the way to the improve your flexibility!
// 'R' returns the 'tRequest' object - object of special type, that contains
// in itself all important data to perform Web API request, and,
// that most importantly, have some methods to change its behaviour!
// See docs for 'tRequest' object and see 'tRequest' methods.
//
// NOTE! Yes, it's very simple. You can just call 'R' method of your
// client object, or call 'R' package level function to get package level
// default client request object, then call all methods to configure request
// sentence and then just call 'IP' or 'IPs' methods and you'll get
// result you want w/o changes behaviour of your 'Client' object!
//
// For example:
// c, err := ipstack.New(...)
// resp, err := c.R().<some_configure_method>(...).IP(1.2.3.4)
//
// And, of course, you can chain all configure methods!
//
// For example:
// _, _ := c.R().<method1>(...).<method2>(...).<method3>(...).IP(1.2.3.4)
//
// WARNING!
// All finishers of 'tRequest' object returns 'tRawResponse' object!
// It means, that you need manually check if any error is occur and
// manually decode the raw JSON response.
// BUT! You can use 'CheckError' and 'DecodeTo' methods of 'tRawResponse'
// class.
// In truth, the 'IP', 'IPs' and 'UpdateMe' methods of 'Client' works that way.
func (c *Client) R() *tRequest {
	if err := c.validate(); err != nil {
		return nil
	}
	return c.baseReq.copy()
}

// 'IP' returns the info about 'ip' as 'Response' object.
// If any error occur, the second return argument will contain error object.
func (c *Client) IP(ip string) (*Response, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	// Save raw response object, check request error
	rr := c.baseReq.IP(ip)
	// Check whether API return an error as encoded JSON
	if err := rr.CheckError(); err != nil {
		return nil, err
	}
	// Try to decode encoded JSON as 'Response' object, check error
	r := Response{}
	if err := rr.DecodeTo(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// 'IPs' returns the info about each requested IP addresses (you can pass
// more than one IP address as arguments) as slice of 'Response' objects.
// You can pass up to 50 IP addresses to the this method.
// If any error occur, the second argument will contain error object.
func (c *Client) IPs(ips ...string) ([]*Response, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	// Save raw response object, check request error
	rr := c.baseReq.IPs(ips...)
	// Check whether API return an error as encoded JSON
	if err := rr.CheckError(); err != nil {
		return nil, err
	}
	// Try to decode encoded JSON as 'Response' object, check error
	r := []*Response{}
	if err := rr.DecodeTo(&r); err != nil {
		return nil, err
	}
	return r, nil
}

// 'Me' fetches the fresh info about your IP address and if this
// operation was successfull, store it as 'Me' field of the current object.
// todo: fix comment
func (c *Client) Me(forceFetch ...bool) (*Response, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	// Should we fetch fresh info, or return cached data
	ff := len(forceFetch) > 0 && forceFetch[0]
	// Return cached data if it is and if fresh fetch do not requested
	if c.me != nil && !ff {
		return c.me, nil
	}
	// Fetch fresh data, check request error
	rr := c.baseReq.Me()
	// Check whether API return an error as encoded JSON
	if err := rr.CheckError(); err != nil {
		return c.me, err
	}
	// Efficency determine where response will store.
	// If 'me' field isn't nil already, decode response to that field.
	// In another way, create a variable to that decode operation will,
	// and in the future, if decode will success, we'll save it to 'me' field.
	r := c.me
	if r == nil {
		r = &Response{}
	}
	// Try to decode encoded JSON as 'Response' object, check error
	if err := rr.DecodeTo(&r); err != nil {
		return c.me, err
	}
	// If this code point is reached, there's no error of JSON decoding
	// and we can safely save 'r' to 'me' field and return it of course.
	c.me = r
	return c.me, nil
}

// 'validate' is the auxiliary method for all public 'Client' methods.
// It checks an internal state of 'Client' object.
// If the internal state is valid, nil is returned, otherwise it returns
// some error caused by incorrect internal state.
//
// In other methods in which this method is called, the returned by this
// method error probably will be returned by the caller too.
//
// Errors and its reasons:
// "Nil client object": You tries to call some method of the nil object
// of client. For example: (*Client)(nil).<some_method>(...).
// "Incorrect internal state": Probably you create 'Client' object directly
// and now you tries to call some method of that object. It's not allowed.
// Use 'New' constructor instead to create a 'Client' instance.
func (c *Client) validate() error {
	if c == nil {
		return fmt.Errorf("Nil client object")
	}
	if c.baseReq == nil {
		return fmt.Errorf("Incorrect internal state")
	}
	return nil
}

// 'applyParams' is the internal private auxiliary function that tries to find
// 'tClientParam' objects in 'params' slice, and if they're found, then tries
// to apply each of them to the current client.
//
// This method calling in 'Client' constructor ('New' function)
// as part of 'Client' initialize process.
func (c *Client) applyParams(params []interface{}) {
	for _, param := range params {
		if param == nil {
			continue
		}
		clientParam, ok := param.(tClientParam)
		if !ok || clientParam == nil {
			continue
		}
		clientParam(c)
	}
}

// 'UseHTTPS' changes the used schema for Web API requests
// to the 'http' or 'https' depends by passed argument as 'is'.
// If 'is' is true, it means 'https'.
// If 'is' is false, it means 'http'.
//
// WARNING! As of 13 Jan 2019, 'https' schema is available only on non-free
// tariff plans! If your tariff plan is 'free' and you'll change to the 'https'
// you probably will get error when you will try to perform any request.
func (r *tRequest) UseHTTPS(is bool) *tRequest {
	if r == nil {
		return nil
	}
	if is == true {
		r.endpoint = cApiEndpointHTTPS
	}
	if is == false {
		r.endpoint = cApiEndpointHTTP
	}
	return r
}

func (r *tRequest) EnableSecuity(is bool) *tRequest {
	if r == nil {
		return nil
	}
	if is == true {
		r.securityEnabled = true
	}

	return r
}

// 'Fields' allows you to specify what kinds of response you want to get
// from ipstack Web API.
//
// You can pass as many fields as you want.
// You can write field names manually or using predefined consts
// started with 'Field...' prefix and described above.
func (r *tRequest) Fields(fields ...string) *tRequest {
	if r == nil {
		return nil
	}
	if len(fields) == 0 {
		return r
	}
	// Save current value of fields only if it isn't empty
	if rf := r.reqArgs.Get("fields"); rf != "" {
		fields = append(fields, rf)
	}
	// Combine all fields with comma separator, encode it and save
	r.reqArgs.Set("fields", strings.Join(fields, ","))
	r.reqArgsBuilt = "?" + r.reqArgs.Encode()
	return r
}

// 'IP' is the one of endpoint to the ipstack Web API that provides
// an info about some one IP address.
// It checks the 'tRequest' object and 'ip' string validities and then
// perform HTTP request to the Web API.
// It returns the 'tResponse' object as it returned from 'do' method.
//
// WARNING! This method have internal check validity of 'ip' address.
// It means, if you pass not valid ip, the method willn't perform any
// HTTP query and just return an error about it.
func (r *tRequest) IP(ip string) *tResponse {
	// Validate 'this' object and arguments
	if err := r.validate(); err != nil {
		return resp(nil, err)
	}
	if ip = strings.TrimSpace(ip); ip == "" {
		return resps(nil, "Empty IP")
	}
	if net.ParseIP(ip) == nil {
		return resps(nil, "Invalid IP (%s)", ip)
	}
	// Make GET request, save result and error of request
	return r.do(ip)
}

// 'IPs' is the one of endpoint to the ipstack Web API that provides
// an info about few IP addresses.
// You can pass up to 50 IP addresses to the this method.
// It checks the 'tRequest' object and each of 'ips' string validities and then
// perform HTTP request to the Web API.
// It returns the 'tResponse' object as it returned from 'do' method.
//
// WARNING! This method have internal check validity of each passed 'ip'
// address. It means, if you pass one or more not valid ip, the method
// will ignore all not valid IP addresses and will perform HTTP query
// only with valid.
func (r *tRequest) IPs(ips ...string) *tResponse {
	// Validate 'this' object and arguments
	if err := r.validate(); err != nil {
		return resp(nil, err)
	}
	if len(ips) == 0 {
		return resps(nil, "No IP passed")
	}
	// Declare slice that will contains only valid ips from 'ips' arg
	// Validate each IP from 'ips' slice. Skip invalid IPs
	validIps := make([]string, 0, len(ips))
	for _, ip := range ips {
		if ip = strings.TrimSpace(ip); ip == "" {
			continue
		}
		if net.ParseIP(ip) == nil {
			continue
		}
		validIps = append(validIps, ip)
	}
	// Check how much valid ips has been passed as args
	if len(validIps) == 0 {
		return resps(nil, "No valid IP passed")
	}
	// Make GET request, save result and error of request
	return r.do(strings.Join(validIps, ","))
}

// 'Me' is the one of endpoint to the ipstack Web API that provides
// an info about the IP address you're owner of.
// It checks the 'tRequest' object validity and then perform HTTP request
// to the Web API.
// It returns the 'tResponse' object as it returned from 'do' method.
func (r *tRequest) Me() *tResponse {
	// Validate 'this' object and arguments
	if err := r.validate(); err != nil {
		return resp(nil, err)
	}
	// Make GET request, return raw response with request error and raw response
	return r.do("check")
}

// 'validate' is auxiliary method for all public 'Client' methods.
// It checks an internal state of 'Client' object.
// If the internal state is valid, nil is returned, otherwise it returns
// some error caused by incorrect internal state.
//
// In other methods in which this method is called, the returned by this
// method error probably will be returned by the caller too.
//
// Errors and its reasons:
// "Nil client object": You tries to call some method of the nil object
// of client. For example: (*Client)(nil).<some_method>(...).
// "Incorrect internal state": Probably you create 'Client' object directly
// and now you tries to call some method of that object. It's not allowed.
// Use 'New' constructor instead to create a 'Client' instance.
func (r *tRequest) validate() error {
	if r == nil {
		return fmt.Errorf("Nil request object")
	}
	if r.client == nil {
		return fmt.Errorf("Nil http.Client object in request")
	}
	return nil
}

// 'do' is the internal private auxiliary method and final part of
// performing Web API request to the ipstack.
// This method generates the full URL string for GET request, then
// perform GET request using generated URL.
// If request was successfull, 'do' tries to read a whole response,
// and save it as 'RawData' field in the returned object.
// If any error occurred, it guarantees that 'RawData' field is empty (nil)
// and 'Error' field contains error object.
//
// So, 'do' work can be split to the subtasks:
// 1. Generate URL
// 2. Perform HTTP/S GET request
// 3. Read a whole JSON response
// *. Check error on each of steps above
func (r *tRequest) do(method string) *tResponse {
	// Make GET request, if any error occur, return it
	url := r.endpoint + method + r.reqArgsBuilt
	if r.securityEnabled {
		url = url + "&" + securityParam
	}
	rr, err := r.client.Get(url)
	if err != nil {
		return resp(nil, err)
	}
	// AFAIK it's not possible, but anyway, if 'Body' of response is nil,
	// return error
	if rr.Body == nil {
		return resps(nil, "Body of GET response is nil")
	}
	// Try to read all as []byte from 'Body' response, and close io.Reader
	// right after reading (w/o deferring because it isn't necessary here)
	b, err := ioutil.ReadAll(rr.Body)
	_ = rr.Body.Close()
	// Check error of reading. If it's not nil, return it
	// Otherwise return readed data
	if err == nil {
		return resp(b, nil)
	}
	return resps(nil, "Error reading Body of GET response (%s)", err)
}

// 'resp' returns the 'tResponse' object created from 'rawData' and 'err'
// objects.
// This function just allocate and create an 'tResponse' instance
// and saves 'rawData' and 'err' objects to the created object.
func resp(rawData []byte, err error) *tResponse {
	return &tResponse{RawData: rawData, Error: err}
}

// 'resps' returns the 'tResponse' object created from 'rawData' and
// 'serr', 'args' objects.
// If 'serr' isn't empty it will be treated as format string to generate
// an error object and 'args' as args for printf-like format string.
// This function just allocate and create an 'tResponse' instance
// and saves 'rawData' and generated error object to the created object.
func resps(rawData []byte, serr string, args ...interface{}) *tResponse {
	r := &tResponse{RawData: rawData}
	if serr != "" {
		r.Error = fmt.Errorf(serr, args...)
	}
	return r
}

// 'copy' is the internal auxiliary method of 'tRequest' object.
// This method creates the full copy of 'tRequest' object and return it.
// It needs to guarantee that applying some changes to the 'tRequest'
// object that will be got by user using 'R' method do not affected
// default client 'tRequest' object.
func (r *tRequest) copy() *tRequest {
	if r == nil {
		return nil
	}
	rr := *r
	return &rr
}

// 'CheckError' checks if Web API response has an error.
// So, if any error has occur when Web API request was performing,
// this method return an occurred error immediately.
// If returned JSON response contains API error, the 'tError' instance
// object will be created and will be stored as 'Error' field in the current
// 'tResponse' object and also returned as object of 'error' interface.
func (r *tResponse) CheckError() error {
	if r == nil {
		return fmt.Errorf("Nil RAW response object")
	}
	// Check if error already occur (request error)
	if r.Error != nil {
		return r.Error
	}
	// Try to decode response JSON to the error object, in which already
	// 'success' field is set to the 'true'
	// If error really occurred, it will be overwritten to the 'false'.
	// It's the fastest way to check error from API I can imagine now
	errApi := tResponseError{Success: true}
	r.Error = json.Unmarshal(r.RawData, &errApi)
	if r.Error != nil {
		return fmt.Errorf("Decode JSON error (%s)", r.Error)
	}
	if errApi.Success == false {
		r.Error = &errApi.Error
		return r.Error
	}
	return nil
}

// 'DecodeTo' tries to unmarshal Web API JSON response stored in the current
// 'tResponse' object as 'RawData' field to the 'i'.
// If any error occurred while trying to decode JSON or already occurred
// ('Error' field isn't empty), the occurred error (from 'Error' filed)
// will be returned.
func (r *tResponse) DecodeTo(i interface{}) error {
	if r == nil {
		return fmt.Errorf("Nil RAW response object")
	}
	if r.Error != nil {
		return r.Error
	}
	if i == nil {
		return fmt.Errorf("Nil destination argument")
	}
	return json.Unmarshal(r.RawData, i)
}

// 'Code' is the auxiliary method of 'tError' class.
// It exists to provide the more convenient way to check error code
// of some error object.
//
// It returns 0 if the current object is nil.
// It might be very useful if you have some 'error' object (err)
// and want to check code of that.
// In that case, you can just write:
// 'APIError(err).Code()' if that call will return 0 it means that:
// - No error occur and 'err' is nil object of 'error' interface
// - Some error occur, but 'err' isn't tError' object (err != nil)
//
// Check 'APIError' docs for details.
func (e *tError) Code() int {
	if e == nil {
		return 0
	}
	return e.RawCode
}

// 'Type' is the auxiliary method of 'tError' class.
// It exists to provide the more convenient way to check error type
// of some error object.
//
// It returns "" (an empty string) if the current object is nil.
// It might be very useful if you have some 'error' object (err)
// and want to check code of that.
// In that case, you can just write:
// 'APIError(err).Typw()' if that call will return empty string it means that:
// - No error occur and 'err' is nil object of 'error' interface
// - Some error occur, but 'err' isn't tError' object (err != nil)
//
// Check 'APIError' docs for details.
func (e *tError) Type() string {
	if e == nil {
		return ""
	}
	return e.RawType
}

// 'Info' is the auxiliary method of 'tError' class.
// It exists to provide the more convenient way to check error info
// of some error object.
//
// It returns "" if the current object is nil.
// It might be very useful if you have some 'error' object (err)
// and want to check code of that.
// In that case, you can just write:
// 'APIError(err).Info()' if that call will return 0 it means that:
// - No error occur and 'err' is nil object of 'error' interface
// - Some error occur, but 'err' isn't tError' object (err != nil)
//
// Check 'APIError' docs for details.
func (e *tError) Info() string {
	if e == nil {
		return ""
	}
	return e.RawInfo
}

// 'Error' implements the 'error' interface for 'tError' class.
// It returns the string that represents a whole Web API error.
// Returned string will be created by the following pattern:
// "[<error_code>]: <error_type> (<error_info>)".
//
// Seee 'tError' docs for more details.
func (e *tError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%d]: %s (%s)", e.RawCode, e.RawType, e.RawInfo)
}

// 'New' creates a new 'Client' instance with specified 'params'.
// When object created and all required prechecks are completed and passed,
// constructor tries to perform test request: Checking IP address of
// machine request made from.
// If it will be successfull, the result will be saved as 'Me' field
// in 'Client' object, and current timestamp will be saved as 'MeTimestamp'.
//
// NOTE: If any error will be occur at the some step, the not nil error object
// will be returned as second return argument.
//
// NOTE: If the second return argument (error object) isn't nil, the first
// argument ('Client' object by pointer) is nil! Always! No exceptions!
//
// Arguments:
// You can pass arguments of some different types in the order you want.
// Arguments will be identified by its type.
// If you will pass argument of unsupported type, it will be ignored.
// if you will pass more than one params of the same type, the all previous
// values will be shadowed by the last value.
// [*] - means that this argument is required.
//
// -- Table of supported argument's types and how they treated are --
//
// [*] {string, []byte} Token. API token that will be used as accessing token
// to perform each request to the ipstack Web API.
// If this argument willn't pass, error will return immediately.
//
// [ ] {http.Client, *http.Client} Golang HTTP client. This object will be
// used to perform each request to the ipstack Web API.
// If this argument willn't pass, the HTTP client with default params
// will be used (see docs for 'http.Client' golang package).
func New(params ...interface{}) (*Client, error) {
	c := &Client{baseReq: &tRequest{reqArgs: url.Values{}}}
	// Apply all params
	c.applyParams(params)
	// Try to extract token from params, if it's not set already, save it
	if c.baseReq.token == "" {
		if c.baseReq.token = extractToken(params); c.baseReq.token == "" {
			return nil, fmt.Errorf("Token argument (string or []byte) is required")
		}
	}
	// Try to extract golang http.Client object from params,
	// if it's not set already, save it
	// Otherwise create the default http.Client object
	if c.baseReq.client == nil {
		if c.baseReq.client = extractHttpClient(params); c.baseReq.client == nil {
			c.baseReq.client = &http.Client{}
		}
	}
	// Set another default values
	if c.baseReq.endpoint == "" {
		c.baseReq.endpoint = cApiEndpointHTTP
	}
	c.baseReq.reqArgs.Set("access_key", c.baseReq.token)
	c.baseReq.reqArgs.Set("output", "json")
	c.baseReq.reqArgsBuilt = "?" + c.baseReq.reqArgs.Encode()
	// Try to perform first query if it's need
	if !c.skipInitFetchMe {
		if _, err := c.Me(); err != nil {
			return nil, fmt.Errorf("Test request error (%s)", err)
		}
	}
	// All good, return 'Client' object and nil as error
	// And also save it as default client if this operation hasn't been
	// performed earlier.
	if DefaultClient == nil {
		DefaultClient = c
	}
	return c, nil
}

// 'extractToken' is auxiliary function for 'Client' constructor
// ('New' package function).
// 'extractToken' tries to find token param in 'params' slice.
// Token type must be 'string' or '[]byte'.
// If 'params' has object with one of these types, it will be treated
// as token (and all spaces in string will be trimmed) and then
// that string will be returned as token.
// If 'params' contains more than one object of these types, the last of them
// will be treated as token.
func extractToken(params []interface{}) (token string) {
	for _, param := range params {
		switch param.(type) {
		case string:
			token = param.(string)
		case []byte:
			token = string(param.([]byte))
		}
	}
	token = strings.TrimSpace(token)
	return token
}

// 'extractHttpClient' is auxiliary function for 'Client' constructor
// ('New' package function).
// 'extractHttpClient' tries to find golang HTTP client object
// in 'params' slice.
// The type of object must be 'http.Client' or '*http.Client'.
// If 'params' has object with one of these types, it will be treated as
// golang HTTP client, and will be returned as 'http.Client' by pointer
// (of course only if it's not nil).
// If 'params' contains more than one object of these types, the last of them
// will be treated as golang HTTP client.
func extractHttpClient(params []interface{}) (client *http.Client) {
	for _, param := range params {
		switch param.(type) {
		case http.Client:
			v := param.(http.Client)
			client = &v
		case *http.Client:
			v := param.(*http.Client)
			if v != nil {
				client = v
			}
		}
	}
	return client
}

// 'ParamToken' creates a parameter for 'Client' constructors that
// specifies the API token, 'Client' object must be created with.
func ParamToken(token string) tClientParam {
	if token = strings.TrimSpace(token); token == "" {
		return nil
	}
	return func(c *Client) {
		if c != nil {
			c.baseReq.token = token
		}
	}
}

// 'ParamDisableFirstMeCall' creates a parameter for 'Client' constructors
// that skips the first internal calling 'Me' method when 'Client' object will
// be created successfully.
//
// NOTE! Yes, when you create a 'Client' object by default it tries to
// perform first test query, and if it was successfull, the 'Client' object
// treates as succeessfully created (and response object is decoded
// and stored as info of the current IP address and will be available
// by calling 'Me' method w/o force fetch).
func ParamDisableFirstMeCall() tClientParam {
	return func(c *Client) {
		if c != nil {
			c.skipInitFetchMe = true
		}
	}
}

// 'ParamUseHTTPS' creates a parameter for 'Client' constructors that
// obliges to always use HTTPS schema instead HTTP.
//
// WARNING! You can use HTTPS schema only if you have non-free ipstack account.
func ParamUseHTTPS(is bool) tClientParam {
	return func(c *Client) {
		if c != nil {
			c.baseReq = c.baseReq.UseHTTPS(is)
		}
	}
}

// 'ParamFields' creates a parameter for 'Client' constructors that allows
// you to specifiy what kinds of response you want to get from ipstack Web API.
//
// You can pass as many fields as you want.
// You can write field names manually or using predefined consts
// started with 'Field...' prefix and described above.
func ParamFields(fields ...string) tClientParam {
	return func(c *Client) {
		if c != nil {
			c.baseReq = c.baseReq.Fields(fields...)
		}
	}
}

// 'ParamEnableSecurity' creates a parameter for 'Client' constructors that
// enables the security module
//
// WARNING! You can use securuty module only if you have non-free ipstack account.
func ParamEnableSecurity(fields ...string) tClientParam {
	return func(c *Client) {
		if c != nil {
			c.baseReq = c.baseReq.EnableSecuity(true)
		}
	}
}

// 'Init' initializes 'DefaultClient' variable.
// This variable represents default client that used by package functions
// 'IP', 'IPs' and 'UpdateMe'.
//
// Thus, you can just once call 'Init' function (initialize default client)
// and then in any place just call 'IP', 'IPs', 'UpdateMe' package level
// functions and accessing to the 'DefaultClient' variable.
// Thus you will have only one 'Client' instance in all your app, and if
// it's enough for you, you do not need to call 'New' directly for creting
// other 'Client' object, and most importantly do not need to worry about
// where to store 'Client' object, or about how to provide access to storing
// object and other architecture solutions.
// Just call 'Init' and then call package level functions.
// It's so easy and simple.
//
// NOTE! You can call this function as much as you want.
// I can't imagine for what you might need this feature, but it's possible.
// In that case, the behaviour is predictable - the previous version
// of default client will be overwritten by new.
// BUT! If any error occurred while trying to reinitialize default client
// (create new object of default client) the previous default client willn't
// be overwritten by nil object and also will be available.
//
// NOTE! See 'New' docs to understand what you must or can pass as arguments.
func Init(params ...interface{}) error {
	cl, err := New(params...)
	if err != nil {
		return err
	}
	DefaultClient = cl
	return nil
}

// 'R' is the same as 'R' of any 'Client' instance but only for default client.
// See docs for 'Client.R' method and 'DefaultClient' variable for details.
func R() *tRequest {
	return DefaultClient.R()
}

// 'IP' is the same as 'IP' of any 'Client' instance
// but only for default client.
// See docs for 'Client.IP' method and 'DefaultClient' variable for details.
func IP(ip string) (*Response, error) {
	if DefaultClient != nil {
		return DefaultClient.IP(ip)
	}
	return nil, fmt.Errorf("DefaultClient client isn't initialized")
}

// 'IPs' is the same as 'IPs' of any 'Client' instance
// but only for default client.
// See docs for 'Client.IPs' method and 'DefaultClient' variable for details.
func IPs(ips ...string) ([]*Response, error) {
	if DefaultClient != nil {
		return DefaultClient.IPs(ips...)
	}
	return nil, fmt.Errorf("DefaultClient client isn't initialized")
}

// 'Me' is the same as 'Me' of any 'Client' instance
// but only for default client.
// See docs for 'Client.Me' method and 'DefaultClient' variable for details.
func Me(forceFetch ...bool) (*Response, error) {
	if DefaultClient != nil {
		return DefaultClient.Me(forceFetch...)
	}
	return nil, fmt.Errorf("DefaultClient client isn't initialized")
}

// 'APIError' tries to cast 'e' object to the 'tError' object.
// 'tError' is the internal private class, represents the some Web API error.
// If 'e' is the object of 'tError' class, it will be returned by pointer,
// otherwise nil is returned.
//
// Because all methods that returns an error object, return object of 'error'
// interface, not an 'tError' instance, in golang you must check if
// some object that implements 'error' interface is really 'tError' instance.
// You can use this function for it and then, for example, if you
// want to check the error code, just call 'Code' method.
// Even if this method will return nil as '*tError' object, method 'Code'
// willn't panic and 0 will be returned by 'Code'.
// Check methods 'Code', 'Type' and 'Info' for details.
func APIError(e error) *tError {
	if e == nil {
		return nil
	}
	if op, ok := e.(*tError); ok {
		return op
	}
	return nil
}
