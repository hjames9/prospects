package common

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"time"
)

type Frequency int

const (
	Always Frequency = 1 << iota
	Hourly
	Daily
	Weekly
	Monthly
	Yearly
	Never
)

func (frequency Frequency) String() string {
	switch frequency {
	case Always:
		return "always"
	case Hourly:
		return "hourly"
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	case Yearly:
		return "yearly"
	case Never:
		return "never"
	default:
		return ""
	}
}

func (frequency Frequency) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	freqStr := frequency.String()
	return encoder.EncodeElement(&freqStr, start)
}

type Url struct {
	Location         string    `xml:"loc"`
	LastModification time.Time `xml:"lastmod,omitempty"`
	ChangeFrequency  Frequency `xml:"changefreq,omitempty"`
	Priority         float32   `xml:"priority,omitempty"`
}

type UrlSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Urls    []Url    `xml:"url"`
}

func (urlset *UrlSet) AddUrl(newUrl Url) error {
	_, err := url.Parse(newUrl.Location)

	if newUrl.Priority < 0.0 || newUrl.Priority > 1.0 {
		err = fmt.Errorf("Invalid priority %f", newUrl.Priority)
	}

	if "" == string(newUrl.ChangeFrequency) {
		err = fmt.Errorf("Invalid change frequency: %d", newUrl.ChangeFrequency)
	}

	if nil == err {
		urlset.Urls = append(urlset.Urls, newUrl)
	}

	return err
}

func (urlset UrlSet) String() string {
	output, err := xml.MarshalIndent(urlset, "", "    ")
	if nil != err {
		return ""
	} else {
		return xml.Header + string(output)
	}
}

type Sitemap struct {
	Location         string    `xml:"loc"`
	LastModification time.Time `xml:"lastmod,omitempty"`
}

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

func (sitemapIdx *SitemapIndex) AddSitemap(sitemap Sitemap) error {
	_, err := url.Parse(sitemap.Location)

	if nil == err {
		sitemapIdx.Sitemaps = append(sitemapIdx.Sitemaps, sitemap)
	}

	return err
}

func (sitemapIdx SitemapIndex) String() string {
	output, err := xml.MarshalIndent(sitemapIdx, "", "    ")
	if nil != err {
		return ""
	} else {
		return xml.Header + string(output)
	}
}
