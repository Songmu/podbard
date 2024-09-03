package cast

import (
	"fmt"
	"net/url"
	"strings"
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
	pd.Generator = fmt.Sprintf("%s powered by %s", generator, pd.Generator)
	pd.AddAuthor(channel.Author, channel.Email)
	pd.AddAtomLink(channel.FeedURL().String())
	pd.Copyright = channel.Copyright
	if img := channel.ImageURL(); img != "" {
		pd.AddImage(img)
	}
	pd.ISubtitle = channel.Description
	pd.AddSummary(channel.Description)

	if len(channel.Categories) == 0 {
		pd.AddCategory("Technology", nil) // default category
	}
	for _, cat := range channel.Categories {
		pd.AddCategory(cat, nil)
	}
	// XXX: pd.IType = "eposodic" // eposodic or serial. eduncan911/podcast does not support this yet

	return &Feed{
		Channel: channel,
		Podcast: pd,
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
		Description: ep.Description,
		Link:        epLink,
		GUID:        epLink,
		ISubtitle:   ep.Description,
		IAuthor:     f.Channel.Author,
	}
	pd := ep.PubDate()
	item.AddPubDate(&pd)
	item.AddSummary(ep.Description)
	item.AddDuration(int64(ep.Audio().Duration))
	// XXX: item.Content = ep.HTML() // <content:encoded> is not supported yet

	audioURL := audioBaseURL.JoinPath(ep.AudioFile)
	encType := podcast.MP3
	if strings.HasSuffix(ep.AudioFile, ".m4a") {
		encType = podcast.M4A
	}
	item.AddEnclosure(audioURL.String(), encType, ep.Audio().FileSize)

	return f.Podcast.AddItem(*item)
}
