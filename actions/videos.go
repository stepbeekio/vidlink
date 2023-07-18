package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/x/responder"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"

	"vidlink/models"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Video)
// DB Table: Plural (videos)
// Resource: Plural (Videos)
// Path: Plural (/videos)
// View Template Folder: Plural (/templates/videos/)

// VideosResource is the resource for the Video model
type VideosResource struct {
	buffalo.Resource
}

// List gets all Videos. This function is mapped to the path
// GET /videos
func (v VideosResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	videos := &models.Video{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Videos from the DB
	if err := q.All(videos); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		c.Set("pagination", q.Paginator)

		c.Set("videos", videos)
		return c.Render(http.StatusOK, r.HTML("videos/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(videos))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(videos))
	}).Respond(c)
}

// Show gets the data for one Video. This function is mapped to
// the path GET /videos/{video_id}
func (v VideosResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Video
	video := &models.Video{}

	// To find the Video the parameter video_id is used.
	if err := tx.Find(video, c.Param("video_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	resolution := c.Param("resolution")
	if resolution == "" {
		resolution = "1280x720"
	}

	c.Set("current_resolution", resolution)
	c.Set("resolutions", []string{"480x270", "640x360", "1280x720", "1920x1080"})

	videoUrl := fmt.Sprintf("%s/%s/quality_%s.m3u8", os.Getenv("SPACES_CDN_URL"), video.ID.String(), "1280x720")

	c.Set("video_link", videoUrl)

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("video", video)

		return c.Render(http.StatusOK, r.HTML("videos/show.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(video))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(video))
	}).Respond(c)
}

// New renders the form for creating a new Video.
// This function is mapped to the path GET /videos/new
func (v VideosResource) New(c buffalo.Context) error {
	c.Set("video", &models.Video{})

	return c.Render(http.StatusOK, r.HTML("videos/new.plush.html"))
}

// Create adds a Video to the DB. This function is mapped to the
// path POST /videos
func (v VideosResource) Create(c buffalo.Context) error {
	// Allocate an empty Video
	video := &models.Video{}

	// Bind video to the html form elements
	if err := c.Bind(video); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(video)
	if err != nil {
		return err
	}

	f, err := c.File("Video")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("File uploaded %s \n", f)
	filename := os.TempDir() + "/" + video.ID.String()
	temp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer temp.Close()

	written, copyErr := io.Copy(temp, f)
	if copyErr != nil {
		return c.Render(500, r.JSON(map[string]string{"error": "Failed to save uploaded file"}))
	}

	fmt.Printf("Copied %s bytes\n", written)

	go models.UploadFileToS3(video, filename, video.ID.String())

	if verrs.HasAny() {
		return responder.Wants("html", func(c buffalo.Context) error {
			// Make the errors available inside the html template
			c.Set("errors", verrs)

			// Render again the new.html template that the user can
			// correct the input.
			c.Set("video", video)

			return c.Render(http.StatusUnprocessableEntity, r.HTML("videos/new.plush.html"))
		}).Wants("json", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
		}).Wants("xml", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
		}).Respond(c)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a success message
		c.Flash().Add("success", T.Translate(c, "video.created.success"))

		// and redirect to the show page
		return c.Redirect(http.StatusSeeOther, "/videos/%v", video.ID)
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusCreated, r.JSON(video))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusCreated, r.XML(video))
	}).Respond(c)
}

// Edit renders a edit form for a Video. This function is
// mapped to the path GET /videos/{video_id}/edit
func (v VideosResource) Edit(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Video
	video := &models.Video{}

	if err := tx.Find(video, c.Param("video_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("video", video)
	return c.Render(http.StatusOK, r.HTML("videos/edit.plush.html"))
}

// Update changes a Video in the DB. This function is mapped to
// the path PUT /videos/{video_id}
func (v VideosResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Video
	video := &models.Video{}

	if err := tx.Find(video, c.Param("video_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Video to the html form elements
	if err := c.Bind(video); err != nil {
		return err
	}

	verrs, err := tx.ValidateAndUpdate(video)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return responder.Wants("html", func(c buffalo.Context) error {
			// Make the errors available inside the html template
			c.Set("errors", verrs)

			// Render again the edit.html template that the user can
			// correct the input.
			c.Set("video", video)

			return c.Render(http.StatusUnprocessableEntity, r.HTML("videos/edit.plush.html"))
		}).Wants("json", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
		}).Wants("xml", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
		}).Respond(c)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a success message
		c.Flash().Add("success", T.Translate(c, "video.updated.success"))

		// and redirect to the show page
		return c.Redirect(http.StatusSeeOther, "/videos/%v", video.ID)
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(video))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(video))
	}).Respond(c)
}

// Destroy deletes a Video from the DB. This function is mapped
// to the path DELETE /videos/{video_id}
func (v VideosResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty Video
	video := &models.Video{}

	// To find the Video the parameter video_id is used.
	if err := tx.Find(video, c.Param("video_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(video); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a flash message
		c.Flash().Add("success", T.Translate(c, "video.destroyed.success"))

		// Redirect to the index page
		return c.Redirect(http.StatusSeeOther, "/videos")
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(video))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(video))
	}).Respond(c)
}
