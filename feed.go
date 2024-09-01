package primcast

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

func NewFeed(channel *ChannelConfig, pubDate time.Time) *Feed {
	pdTmp := podcast.New(channel.Title, channel.Link, channel.Description, &pubDate, &pubDate)

	pd := &pdTmp
	pd.Language = channel.Language
	pd.Generator = fmt.Sprintf("github.com/Songmu/primcast %s powered by %s", version, pd.Generator)
	pd.AddAuthor(channel.Author, channel.Email)
	pd.AddAtomLink(channel.FeedURL())
	pd.Copyright = channel.Copyright
	pd.AddImage(channel.ImageURL())

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

func (f *Feed) AddEpisode(ep *Episode, audioBucketURL string) (int, error) {
	epLink, err := url.JoinPath(f.Channel.Link, episodeDir, ep.Name)
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

	audioBaseURL := audioBucketURL
	if audioBaseURL == "" {
		// f.Channel.Link has been already validated above
		audioBaseURL, _ = url.JoinPath(f.Channel.Link, audioDir)
	}
	audioURL, err := url.JoinPath(audioBaseURL, ep.AudioFile)
	if err != nil {
		return 0, err
	}
	encType := podcast.MP3
	if strings.HasSuffix(ep.AudioFile, ".m4a") {
		encType = podcast.M4A
	}
	item.AddEnclosure(audioURL, encType, ep.Audio().FileSize)

	return f.Podcast.AddItem(*item)
}
