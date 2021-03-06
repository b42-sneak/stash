package manager

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stashapp/stash/pkg/database"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/manager/config"
	"github.com/stashapp/stash/pkg/manager/jsonschema"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

type ImportTask struct {
	Mappings *jsonschema.Mappings
	Scraped  []jsonschema.ScrapedItem
}

func (t *ImportTask) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	t.Mappings, _ = instance.JSON.getMappings()
	if t.Mappings == nil {
		logger.Error("missing mappings json")
		return
	}
	scraped, _ := instance.JSON.getScraped()
	if scraped == nil {
		logger.Warn("missing scraped json")
	}
	t.Scraped = scraped

	err := database.Reset(config.GetDatabasePath())

	if err != nil {
		logger.Errorf("Error resetting database: %s", err.Error())
		return
	}

	ctx := context.TODO()

	t.ImportPerformers(ctx)
	t.ImportStudios(ctx)
	t.ImportGalleries(ctx)
	t.ImportTags(ctx)

	t.ImportScrapedItems(ctx)
	t.ImportScenes(ctx)
}

func (t *ImportTask) ImportPerformers(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewPerformerQueryBuilder()

	for i, mappingJSON := range t.Mappings.Performers {
		index := i + 1
		performerJSON, err := instance.JSON.getPerformer(mappingJSON.Checksum)
		if err != nil {
			logger.Errorf("[performers] failed to read json: %s", err.Error())
			continue
		}
		if mappingJSON.Checksum == "" || mappingJSON.Name == "" || performerJSON == nil {
			return
		}

		logger.Progressf("[performers] %d of %d", index, len(t.Mappings.Performers))

		// generate checksum from performer name rather than image
		checksum := utils.MD5FromString(performerJSON.Name)

		// Process the base 64 encoded image string
		_, imageData, err := utils.ProcessBase64Image(performerJSON.Image)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[performers] <%s> invalid image: %s", mappingJSON.Checksum, err.Error())
			return
		}

		// Populate a new performer from the input
		newPerformer := models.Performer{
			Image:     imageData,
			Checksum:  checksum,
			Favorite:  sql.NullBool{Bool: performerJSON.Favorite, Valid: true},
			CreatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(performerJSON.CreatedAt)},
			UpdatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(performerJSON.UpdatedAt)},
		}

		if performerJSON.Name != "" {
			newPerformer.Name = sql.NullString{String: performerJSON.Name, Valid: true}
		}
		if performerJSON.URL != "" {
			newPerformer.URL = sql.NullString{String: performerJSON.URL, Valid: true}
		}
		if performerJSON.Birthdate != "" {
			newPerformer.Birthdate = models.SQLiteDate{String: performerJSON.Birthdate, Valid: true}
		}
		if performerJSON.Ethnicity != "" {
			newPerformer.Ethnicity = sql.NullString{String: performerJSON.Ethnicity, Valid: true}
		}
		if performerJSON.Country != "" {
			newPerformer.Country = sql.NullString{String: performerJSON.Country, Valid: true}
		}
		if performerJSON.EyeColor != "" {
			newPerformer.EyeColor = sql.NullString{String: performerJSON.EyeColor, Valid: true}
		}
		if performerJSON.Height != "" {
			newPerformer.Height = sql.NullString{String: performerJSON.Height, Valid: true}
		}
		if performerJSON.Measurements != "" {
			newPerformer.Measurements = sql.NullString{String: performerJSON.Measurements, Valid: true}
		}
		if performerJSON.FakeTits != "" {
			newPerformer.FakeTits = sql.NullString{String: performerJSON.FakeTits, Valid: true}
		}
		if performerJSON.CareerLength != "" {
			newPerformer.CareerLength = sql.NullString{String: performerJSON.CareerLength, Valid: true}
		}
		if performerJSON.Tattoos != "" {
			newPerformer.Tattoos = sql.NullString{String: performerJSON.Tattoos, Valid: true}
		}
		if performerJSON.Piercings != "" {
			newPerformer.Piercings = sql.NullString{String: performerJSON.Piercings, Valid: true}
		}
		if performerJSON.Aliases != "" {
			newPerformer.Aliases = sql.NullString{String: performerJSON.Aliases, Valid: true}
		}
		if performerJSON.Twitter != "" {
			newPerformer.Twitter = sql.NullString{String: performerJSON.Twitter, Valid: true}
		}
		if performerJSON.Instagram != "" {
			newPerformer.Instagram = sql.NullString{String: performerJSON.Instagram, Valid: true}
		}

		_, err = qb.Create(newPerformer, tx)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[performers] <%s> failed to create: %s", mappingJSON.Checksum, err.Error())
			return
		}
	}

	logger.Info("[performers] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[performers] import failed to commit: %s", err.Error())
	}
	logger.Info("[performers] import complete")
}

func (t *ImportTask) ImportStudios(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewStudioQueryBuilder()

	for i, mappingJSON := range t.Mappings.Studios {
		index := i + 1
		studioJSON, err := instance.JSON.getStudio(mappingJSON.Checksum)
		if err != nil {
			logger.Errorf("[studios] failed to read json: %s", err.Error())
			continue
		}
		if mappingJSON.Checksum == "" || mappingJSON.Name == "" || studioJSON == nil {
			return
		}

		logger.Progressf("[studios] %d of %d", index, len(t.Mappings.Studios))

		// generate checksum from studio name rather than image
		checksum := utils.MD5FromString(studioJSON.Name)

		// Process the base 64 encoded image string
		_, imageData, err := utils.ProcessBase64Image(studioJSON.Image)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[studios] <%s> invalid image: %s", mappingJSON.Checksum, err.Error())
			return
		}

		// Populate a new studio from the input
		newStudio := models.Studio{
			Image:     imageData,
			Checksum:  checksum,
			Name:      sql.NullString{String: studioJSON.Name, Valid: true},
			URL:       sql.NullString{String: studioJSON.URL, Valid: true},
			CreatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(studioJSON.CreatedAt)},
			UpdatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(studioJSON.UpdatedAt)},
		}

		_, err = qb.Create(newStudio, tx)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[studios] <%s> failed to create: %s", mappingJSON.Checksum, err.Error())
			return
		}
	}

	logger.Info("[studios] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[studios] import failed to commit: %s", err.Error())
	}
	logger.Info("[studios] import complete")
}

func (t *ImportTask) ImportGalleries(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewGalleryQueryBuilder()

	for i, mappingJSON := range t.Mappings.Galleries {
		index := i + 1
		if mappingJSON.Checksum == "" || mappingJSON.Path == "" {
			return
		}

		logger.Progressf("[galleries] %d of %d", index, len(t.Mappings.Galleries))

		// Populate a new gallery from the input
		currentTime := time.Now()
		newGallery := models.Gallery{
			Checksum:  mappingJSON.Checksum,
			Path:      mappingJSON.Path,
			CreatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
			UpdatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
		}

		_, err := qb.Create(newGallery, tx)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[galleries] <%s> failed to create: %s", mappingJSON.Checksum, err.Error())
			return
		}
	}

	logger.Info("[galleries] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[galleries] import failed to commit: %s", err.Error())
	}
	logger.Info("[galleries] import complete")
}

func (t *ImportTask) ImportTags(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewTagQueryBuilder()

	var tagNames []string

	for i, mappingJSON := range t.Mappings.Scenes {
		index := i + 1
		if mappingJSON.Checksum == "" || mappingJSON.Path == "" {
			_ = tx.Rollback()
			logger.Warn("[tags] scene mapping without checksum or path: ", mappingJSON)
			return
		}

		logger.Progressf("[tags] %d of %d scenes", index, len(t.Mappings.Scenes))

		sceneJSON, err := instance.JSON.getScene(mappingJSON.Checksum)
		if err != nil {
			logger.Infof("[tags] <%s> json parse failure: %s", mappingJSON.Checksum, err.Error())
		}
		// Return early if we are missing a json file.
		if sceneJSON == nil {
			continue
		}

		// Get the tags from the tags json if we have it
		if len(sceneJSON.Tags) > 0 {
			tagNames = append(tagNames, sceneJSON.Tags...)
		}

		// Get the tags from the markers if we have marker json
		if len(sceneJSON.Markers) == 0 {
			continue
		}
		for _, markerJSON := range sceneJSON.Markers {
			if markerJSON.PrimaryTag != "" {
				tagNames = append(tagNames, markerJSON.PrimaryTag)
			}
			if len(markerJSON.Tags) > 0 {
				tagNames = append(tagNames, markerJSON.Tags...)
			}
		}
	}

	uniqueTagNames := t.getUnique(tagNames)
	for _, tagName := range uniqueTagNames {
		currentTime := time.Now()
		newTag := models.Tag{
			Name:      tagName,
			CreatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
			UpdatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
		}

		_, err := qb.Create(newTag, tx)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[tags] <%s> failed to create: %s", tagName, err.Error())
			return
		}
	}

	logger.Info("[tags] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[tags] import failed to commit: %s", err.Error())
	}
	logger.Info("[tags] import complete")
}

func (t *ImportTask) ImportScrapedItems(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewScrapedItemQueryBuilder()
	sqb := models.NewStudioQueryBuilder()
	currentTime := time.Now()

	for i, mappingJSON := range t.Scraped {
		index := i + 1
		logger.Progressf("[scraped sites] %d of %d", index, len(t.Mappings.Scenes))

		newScrapedItem := models.ScrapedItem{
			Title:           sql.NullString{String: mappingJSON.Title, Valid: true},
			Description:     sql.NullString{String: mappingJSON.Description, Valid: true},
			URL:             sql.NullString{String: mappingJSON.URL, Valid: true},
			Date:            models.SQLiteDate{String: mappingJSON.Date, Valid: true},
			Rating:          sql.NullString{String: mappingJSON.Rating, Valid: true},
			Tags:            sql.NullString{String: mappingJSON.Tags, Valid: true},
			Models:          sql.NullString{String: mappingJSON.Models, Valid: true},
			Episode:         sql.NullInt64{Int64: int64(mappingJSON.Episode), Valid: true},
			GalleryFilename: sql.NullString{String: mappingJSON.GalleryFilename, Valid: true},
			GalleryURL:      sql.NullString{String: mappingJSON.GalleryURL, Valid: true},
			VideoFilename:   sql.NullString{String: mappingJSON.VideoFilename, Valid: true},
			VideoURL:        sql.NullString{String: mappingJSON.VideoURL, Valid: true},
			CreatedAt:       models.SQLiteTimestamp{Timestamp: currentTime},
			UpdatedAt:       models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(mappingJSON.UpdatedAt)},
		}

		studio, err := sqb.FindByName(mappingJSON.Studio, tx)
		if err != nil {
			logger.Errorf("[scraped sites] failed to fetch studio: %s", err.Error())
		}
		if studio != nil {
			newScrapedItem.StudioID = sql.NullInt64{Int64: int64(studio.ID), Valid: true}
		}

		_, err = qb.Create(newScrapedItem, tx)
		if err != nil {
			logger.Errorf("[scraped sites] <%s> failed to create: %s", newScrapedItem.Title.String, err.Error())
		}
	}

	logger.Info("[scraped sites] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[scraped sites] import failed to commit: %s", err.Error())
	}
	logger.Info("[scraped sites] import complete")
}

func (t *ImportTask) ImportScenes(ctx context.Context) {
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewSceneQueryBuilder()
	jqb := models.NewJoinsQueryBuilder()

	for i, mappingJSON := range t.Mappings.Scenes {
		index := i + 1
		if mappingJSON.Checksum == "" || mappingJSON.Path == "" {
			_ = tx.Rollback()
			logger.Warn("[scenes] scene mapping without checksum or path: ", mappingJSON)
			return
		}

		logger.Progressf("[scenes] %d of %d", index, len(t.Mappings.Scenes))

		newScene := models.Scene{
			Checksum: mappingJSON.Checksum,
			Path:     mappingJSON.Path,
		}

		sceneJSON, err := instance.JSON.getScene(mappingJSON.Checksum)
		if err != nil {
			logger.Infof("[scenes] <%s> json parse failure: %s", mappingJSON.Checksum, err.Error())
			continue
		}

		// Populate scene fields
		if sceneJSON != nil {
			if sceneJSON.Title != "" {
				newScene.Title = sql.NullString{String: sceneJSON.Title, Valid: true}
			}
			if sceneJSON.Details != "" {
				newScene.Details = sql.NullString{String: sceneJSON.Details, Valid: true}
			}
			if sceneJSON.URL != "" {
				newScene.URL = sql.NullString{String: sceneJSON.URL, Valid: true}
			}
			if sceneJSON.Date != "" {
				newScene.Date = models.SQLiteDate{String: sceneJSON.Date, Valid: true}
			}
			if sceneJSON.Rating != 0 {
				newScene.Rating = sql.NullInt64{Int64: int64(sceneJSON.Rating), Valid: true}
			}
			newScene.CreatedAt = models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(sceneJSON.CreatedAt)}
			newScene.UpdatedAt = models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(sceneJSON.UpdatedAt)}

			if sceneJSON.File != nil {
				if sceneJSON.File.Size != "" {
					newScene.Size = sql.NullString{String: sceneJSON.File.Size, Valid: true}
				}
				if sceneJSON.File.Duration != "" {
					duration, _ := strconv.ParseFloat(sceneJSON.File.Duration, 64)
					newScene.Duration = sql.NullFloat64{Float64: duration, Valid: true}
				}
				if sceneJSON.File.VideoCodec != "" {
					newScene.VideoCodec = sql.NullString{String: sceneJSON.File.VideoCodec, Valid: true}
				}
				if sceneJSON.File.AudioCodec != "" {
					newScene.AudioCodec = sql.NullString{String: sceneJSON.File.AudioCodec, Valid: true}
				}
				if sceneJSON.File.Width != 0 {
					newScene.Width = sql.NullInt64{Int64: int64(sceneJSON.File.Width), Valid: true}
				}
				if sceneJSON.File.Height != 0 {
					newScene.Height = sql.NullInt64{Int64: int64(sceneJSON.File.Height), Valid: true}
				}
				if sceneJSON.File.Framerate != "" {
					framerate, _ := strconv.ParseFloat(sceneJSON.File.Framerate, 64)
					newScene.Framerate = sql.NullFloat64{Float64: framerate, Valid: true}
				}
				if sceneJSON.File.Bitrate != 0 {
					newScene.Bitrate = sql.NullInt64{Int64: int64(sceneJSON.File.Bitrate), Valid: true}
				}
			} else {
				// TODO: Get FFMPEG data?
			}
		}

		// Populate the studio ID
		if sceneJSON.Studio != "" {
			sqb := models.NewStudioQueryBuilder()
			studio, err := sqb.FindByName(sceneJSON.Studio, tx)
			if err != nil {
				logger.Warnf("[scenes] studio <%s> does not exist: %s", sceneJSON.Studio, err.Error())
			} else {
				newScene.StudioID = sql.NullInt64{Int64: int64(studio.ID), Valid: true}
			}
		}

		// Create the scene in the DB
		scene, err := qb.Create(newScene, tx)
		if err != nil {
			_ = tx.Rollback()
			logger.Errorf("[scenes] <%s> failed to create: %s", mappingJSON.Checksum, err.Error())
			return
		}
		if scene.ID == 0 {
			_ = tx.Rollback()
			logger.Errorf("[scenes] <%s> invalid id after scene creation", mappingJSON.Checksum)
			return
		}

		// Relate the scene to the gallery
		if sceneJSON.Gallery != "" {
			gqb := models.NewGalleryQueryBuilder()
			gallery, err := gqb.FindByChecksum(sceneJSON.Gallery, tx)
			if err != nil {
				logger.Warnf("[scenes] gallery <%s> does not exist: %s", sceneJSON.Gallery, err.Error())
			} else {
				gallery.SceneID = sql.NullInt64{Int64: int64(scene.ID), Valid: true}
				_, err := gqb.Update(*gallery, tx)
				if err != nil {
					logger.Errorf("[scenes] <%s> failed to update gallery: %s", scene.Checksum, err.Error())
				}
			}
		}

		// Relate the scene to the performers
		if len(sceneJSON.Performers) > 0 {
			performers, err := t.getPerformers(sceneJSON.Performers, tx)
			if err != nil {
				logger.Warnf("[scenes] <%s> failed to fetch performers: %s", scene.Checksum, err.Error())
			} else {
				var performerJoins []models.PerformersScenes
				for _, performer := range performers {
					join := models.PerformersScenes{
						PerformerID: performer.ID,
						SceneID:     scene.ID,
					}
					performerJoins = append(performerJoins, join)
				}
				if err := jqb.CreatePerformersScenes(performerJoins, tx); err != nil {
					logger.Errorf("[scenes] <%s> failed to associate performers: %s", scene.Checksum, err.Error())
				}
			}
		}

		// Relate the scene to the tags
		if len(sceneJSON.Tags) > 0 {
			tags, err := t.getTags(scene.Checksum, sceneJSON.Tags, tx)
			if err != nil {
				logger.Warnf("[scenes] <%s> failed to fetch tags: %s", scene.Checksum, err.Error())
			} else {
				var tagJoins []models.ScenesTags
				for _, tag := range tags {
					join := models.ScenesTags{
						SceneID: scene.ID,
						TagID:   tag.ID,
					}
					tagJoins = append(tagJoins, join)
				}
				if err := jqb.CreateScenesTags(tagJoins, tx); err != nil {
					logger.Errorf("[scenes] <%s> failed to associate tags: %s", scene.Checksum, err.Error())
				}
			}
		}

		// Relate the scene to the scene markers
		if len(sceneJSON.Markers) > 0 {
			smqb := models.NewSceneMarkerQueryBuilder()
			tqb := models.NewTagQueryBuilder()
			for _, marker := range sceneJSON.Markers {
				seconds, _ := strconv.ParseFloat(marker.Seconds, 64)
				newSceneMarker := models.SceneMarker{
					Title:     marker.Title,
					Seconds:   seconds,
					SceneID:   sql.NullInt64{Int64: int64(scene.ID), Valid: true},
					CreatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(marker.CreatedAt)},
					UpdatedAt: models.SQLiteTimestamp{Timestamp: t.getTimeFromJSONTime(marker.UpdatedAt)},
				}

				primaryTag, err := tqb.FindByName(marker.PrimaryTag, tx)
				if err != nil {
					logger.Errorf("[scenes] <%s> failed to find primary tag for marker: %s", scene.Checksum, err.Error())
				} else {
					newSceneMarker.PrimaryTagID = primaryTag.ID
				}

				// Create the scene marker in the DB
				sceneMarker, err := smqb.Create(newSceneMarker, tx)
				if err != nil {
					logger.Warnf("[scenes] <%s> failed to create scene marker: %s", scene.Checksum, err.Error())
					continue
				}
				if sceneMarker.ID == 0 {
					logger.Warnf("[scenes] <%s> invalid scene marker id after scene marker creation", scene.Checksum)
					continue
				}

				// Get the scene marker tags and create the joins
				tags, err := t.getTags(scene.Checksum, marker.Tags, tx)
				if err != nil {
					logger.Warnf("[scenes] <%s> failed to fetch scene marker tags: %s", scene.Checksum, err.Error())
				} else {
					var tagJoins []models.SceneMarkersTags
					for _, tag := range tags {
						join := models.SceneMarkersTags{
							SceneMarkerID: sceneMarker.ID,
							TagID:         tag.ID,
						}
						tagJoins = append(tagJoins, join)
					}
					if err := jqb.CreateSceneMarkersTags(tagJoins, tx); err != nil {
						logger.Errorf("[scenes] <%s> failed to associate scene marker tags: %s", scene.Checksum, err.Error())
					}
				}
			}
		}
	}

	logger.Info("[scenes] importing")
	if err := tx.Commit(); err != nil {
		logger.Errorf("[scenes] import failed to commit: %s", err.Error())
	}
	logger.Info("[scenes] import complete")
}

func (t *ImportTask) getPerformers(names []string, tx *sqlx.Tx) ([]*models.Performer, error) {
	pqb := models.NewPerformerQueryBuilder()
	performers, err := pqb.FindByNames(names, tx)
	if err != nil {
		return nil, err
	}

	var pluckedNames []string
	for _, performer := range performers {
		if !performer.Name.Valid {
			continue
		}
		pluckedNames = append(pluckedNames, performer.Name.String)
	}

	missingPerformers := utils.StrFilter(names, func(name string) bool {
		return !utils.StrInclude(pluckedNames, name)
	})

	for _, missingPerformer := range missingPerformers {
		logger.Warnf("[scenes] performer %s does not exist", missingPerformer)
	}

	return performers, nil
}

func (t *ImportTask) getTags(sceneChecksum string, names []string, tx *sqlx.Tx) ([]*models.Tag, error) {
	tqb := models.NewTagQueryBuilder()
	tags, err := tqb.FindByNames(names, tx)
	if err != nil {
		return nil, err
	}

	var pluckedNames []string
	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		pluckedNames = append(pluckedNames, tag.Name)
	}

	missingTags := utils.StrFilter(names, func(name string) bool {
		return !utils.StrInclude(pluckedNames, name)
	})

	for _, missingTag := range missingTags {
		logger.Warnf("[scenes] <%s> tag %s does not exist", sceneChecksum, missingTag)
	}

	return tags, nil
}

// https://www.reddit.com/r/golang/comments/5ia523/idiomatic_way_to_remove_duplicates_in_a_slice/db6qa2e
func (t *ImportTask) getUnique(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

var currentLocation = time.Now().Location()

func (t *ImportTask) getTimeFromJSONTime(jsonTime models.JSONTime) time.Time {
	if currentLocation != nil {
		if jsonTime.IsZero() {
			return time.Now().In(currentLocation)
		} else {
			return jsonTime.Time.In(currentLocation)
		}
	} else {
		if jsonTime.IsZero() {
			return time.Now()
		} else {
			return jsonTime.Time
		}
	}
}
