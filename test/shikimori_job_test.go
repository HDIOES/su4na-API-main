package test

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/HDIOES/cpa-backend/models"
	"github.com/pkg/errors"

	"github.com/HDIOES/cpa-backend/integration"
	"github.com/HDIOES/cpa-backend/rest/util"
	_ "github.com/lib/pq"

	gock "gopkg.in/h2non/gock.v1"
)

//TestSimple function
func TestShikimoriJobSuccess(t *testing.T) {
	diContainer.Invoke(func(configuration *util.Configuration, job *integration.ShikimoriJob, newDao *models.NewDAO, animeDao *models.AnimeDAO, genreDao *models.GenreDAO, studioDao *models.StudioDAO) {
		if err := clearDb(newDao, animeDao, genreDao, studioDao); err != nil {
			markAsFailAndAbortNow(t, errors.Wrap(err, ""))
		}
		defer gock.Off()
		animesData, err := ioutil.ReadFile("mock/shikimori_animes_success.json")
		if err != nil {
			markAsFailAndAbortNow(t, errors.Wrap(err, ""))
		}
		gock.New(configuration.ShikimoriURL).
			Get(configuration.ShikimoriAnimeSearchURL).
			MatchParam("page", "1").
			MatchParam("limit", "50").
			Reply(200).
			JSON(animesData)

		genresData, err := ioutil.ReadFile("mock/shikimori_genres_success.json")
		if err != nil {
			markAsFailAndAbortNow(t, errors.Wrap(err, ""))
		}
		gock.New(configuration.ShikimoriURL).
			Get(configuration.ShikimoriGenreURL).
			Reply(200).
			JSON(genresData)

		studiosData, err := ioutil.ReadFile("mock/shikimori_studios_success.json")
		if err != nil {
			markAsFailAndAbortNow(t, errors.Wrap(err, ""))
		}
		gock.New(configuration.ShikimoriURL).
			Get(configuration.ShikimoriStudioURL).
			Reply(200).
			JSON(studiosData)

		oneAnimeData, err := ioutil.ReadFile("mock/one_anime_shikimori_success.json")
		if err != nil {
			markAsFailAndAbortNow(t, errors.Wrap(err, ""))
		}
		gock.New(configuration.ShikimoriURL).
			Get(configuration.ShikimoriAnimeSearchURL+"/").
			PathParam("animes", "5114").
			Reply(200).
			JSON(oneAnimeData)

		job.Run()

		//asserts
		anime := integration.Anime{}
		genres := []integration.Genre{}
		studios := []integration.Studio{}
		if unmarshalAnimeErr := json.Unmarshal(oneAnimeData, &anime); unmarshalAnimeErr != nil {
			markAsFailAndAbortNow(t, errors.Wrap(unmarshalAnimeErr, ""))
		}
		if unmarshalGenresErr := json.Unmarshal(genresData, &genres); unmarshalGenresErr != nil {
			markAsFailAndAbortNow(t, errors.Wrap(unmarshalGenresErr, ""))
		}
		if unmarshalStudioErr := json.Unmarshal(studiosData, &studios); unmarshalStudioErr != nil {
			markAsFailAndAbortNow(t, errors.Wrap(unmarshalStudioErr, ""))
		}
		for _, g := range genres {
			genreDto, genreDtoErr := genreDao.FindByExternalID(strconv.FormatInt(*g.ID, 10))
			if genreDtoErr != nil {
				markAsFailAndAbortNow(t, errors.Wrap(genreDtoErr, ""))
			}
			abortIfFail(t, assert.Equal(t, strconv.FormatInt(*g.ID, 10), genreDto.ExternalID))
			abortIfFail(t, EqualStringValues(t, g.Kind, genreDto.Kind))
			abortIfFail(t, EqualStringValues(t, g.Name, genreDto.Name))
			abortIfFail(t, EqualStringValues(t, g.Russian, genreDto.Russian))
		}
		for _, s := range studios {
			studioDto, studioDtoErr := studioDao.FindByExternalID(strconv.FormatInt(*s.ID, 10))
			if studioDtoErr != nil {
				markAsFailAndAbortNow(t, errors.Wrap(studioDtoErr, ""))
			}
			abortIfFail(t, assert.Equal(t, strconv.FormatInt(*s.ID, 10), studioDto.ExternalID))
			abortIfFail(t, EqualStringValues(t, s.FilteredName, studioDto.FilteredStudioName))
			abortIfFail(t, EqualStringValues(t, s.Image, studioDto.ImageURL))
			abortIfFail(t, EqualStringValues(t, s.Name, studioDto.Name))
			abortIfFail(t, EqualBoolValues(t, s.Real, studioDto.IsReal))
		}
	})
}
