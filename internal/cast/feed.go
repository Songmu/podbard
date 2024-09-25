package cast

import (
	"fmt"
	"net/url"
	"time"

	"github.com/eduncan911/podcast"
)

type Feed struct {
	Channel *ChannelConfig
	Podcast *podcast.Podcast
}

func NewFeed(generator string, channel *ChannelConfig, pubDate, lastBuildDate time.Time) *Feed {
	pdTmp := podcast.New(
		channel.Title, channel.Link.String(), channel.Description, &pubDate, &lastBuildDate)

	pd := &pdTmp
	pd.Language = channel.Language.String()
	pd.AddAuthor(channel.Author, "")
	pd.IAuthor = channel.Author
	pd.IOwner = &podcast.Author{
		Name:  channel.Author,
		Email: channel.Email,
	}
	pd.AddAtomLink(channel.FeedURL().String())
	if img := channel.ImageURL(); img != "" {
		pd.AddImage(img)
	}
	if len(channel.Categories) == 0 {
		pd.AddCategory("Technology", nil) // default category
	}
	for _, cat := range channel.Categories {
		pd.AddCategory(cat, nil)
	}
	pd.IExplicit = fmt.Sprintf("%t", channel.Explicit)
	pd.Generator = generator
	pd.Copyright = channel.Copyright
	if channel.Copyright != "" {
		pd.Copyright = fmt.Sprintf("&#xA9; 2024 %s", channel.Author) // XXX: year is hardcoded
	}

	// deprecated but used tags
	pd.ISubtitle = channel.Description

	// XXX: pd.IType = "eposodic" // <itunes:type> eposodic or serial. eduncan911/podcast does not support this

	return &Feed{
		Channel: channel,
		Podcast: pd,
	}
}

func podcastEnclosureType(mediaType MediaType) (podcast.EnclosureType, bool) {
	switch mediaType {
	case MP3:
		return podcast.MP3, true
	case M4A:
		return podcast.M4A, true
	default:
		return 0, false
	}
}

func (f *Feed) AddEpisode(ep *Episode, audioBaseURL *url.URL) (int, error) {
	epLink, err := url.JoinPath(f.Channel.Link.String(), episodeDir, ep.Slug)
	if err != nil {
		return 0, err
	}
	epLink += "/"
	item := &podcast.Item{
		Title:       ep.Title,
		Description: buildCData(ep.Body),
		Link:        epLink,
		GUID:        epLink,
		IExplicit:   fmt.Sprintf("%t", f.Channel.Explicit),
		// don't use `item.AddDuration(d int64)`. It converts duration to string like "53:12",
		// but just use seconds is recommended by Apple.
		IDuration: fmt.Sprintf("%d", ep.Audio().Duration),
	}
	pd := ep.PubDate()
	item.AddPubDate(&pd)
	if img := f.Channel.ImageURL(); img != "" {
		item.AddImage(img)
	}

	// deprecated but used tags
	item.ISubtitle = ep.Subtitle
	item.IAuthor = f.Channel.Author

	// XXX: item.IEpisodeType = "full" // <itunes:episodeType> full, trailer or bonus.
	//                                 // eduncan911/podcast does not support this
	// XXX: item.Content = ep.HTML() // <content:encoded> is not supported yet

	audioURL := audioBaseURL.JoinPath(ep.AudioFile)
	encType, ok := podcastEnclosureType(ep.Audio().mediaType)
	if !ok {
		return 0, fmt.Errorf("unsupported media type: %s", ep.AudioFile)
	}
	item.AddEnclosure(audioURL.String(), encType, ep.Audio().FileSize)

	return f.Podcast.AddItem(*item)
}

func buildCData(data string) string {
	return "<![CDATA[" + data + "]]>"
}
