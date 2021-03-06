package apiv1

import (
	"github.com/getfider/fider/app/actions"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/web"
	"github.com/getfider/fider/app/tasks"
)

// SearchPosts return existing posts based on search criteria
func SearchPosts() web.HandlerFunc {
	return func(c web.Context) error {
		posts, err := c.Services().Posts.Search(
			c.QueryParam("q"),
			c.QueryParam("f"),
			c.QueryParam("l"),
			c.QueryParamAsArray("t"),
		)
		if err != nil {
			return c.Failure(err)
		}
		return c.Ok(posts)
	}
}

// CreatePost creates a new post on current tenant
func CreatePost() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.CreateNewPost)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		posts := c.Services().Posts
		post, err := posts.Add(input.Model.Title, input.Model.Description)
		if err != nil {
			return c.Failure(err)
		}

		if err := posts.AddVote(post, c.User()); err != nil {
			return c.Failure(err)
		}

		c.Enqueue(tasks.NotifyAboutNewPost(post))

		return c.Ok(post)
	}
}

// UpdatePost updates an existing post of current tenant
func UpdatePost() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.UpdatePost)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		_, err := c.Services().Posts.Update(input.Post, input.Model.Title, input.Model.Description)
		if err != nil {
			return c.Failure(err)
		}

		return c.Ok(web.Map{})
	}
}

// SetResponse changes current post staff response
func SetResponse() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.SetResponse)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		post, err := c.Services().Posts.GetByNumber(input.Model.Number)
		if err != nil {
			return c.Failure(err)
		}

		prevStatus := post.Status
		if input.Model.Status == models.PostDuplicate {
			err = c.Services().Posts.MarkAsDuplicate(post, input.Original)
		} else {
			err = c.Services().Posts.SetResponse(post, input.Model.Text, input.Model.Status)
		}
		if err != nil {
			return c.Failure(err)
		}

		c.Enqueue(tasks.NotifyAboutStatusChange(post, prevStatus))

		return c.Ok(web.Map{})
	}
}

// DeletePost deletes an existing post of current tenant
func DeletePost() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.DeletePost)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		err := c.Services().Posts.SetResponse(input.Post, input.Model.Text, models.PostDeleted)
		if err != nil {
			return c.Failure(err)
		}

		return c.Ok(web.Map{})
	}
}

// ListComments returns a list of all comments of a post
func ListComments() web.HandlerFunc {
	return func(c web.Context) error {
		number, err := c.ParamAsInt("number")
		if err != nil {
			return c.Failure(err)
		}

		post, err := c.Services().Posts.GetByNumber(number)
		if err != nil {
			return c.Failure(err)
		}

		comments, err := c.Services().Posts.GetCommentsByPost(post)
		if err != nil {
			return c.Failure(err)
		}

		return c.Ok(comments)
	}
}

// PostComment creates a new comment on given post
func PostComment() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.AddNewComment)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		post, err := c.Services().Posts.GetByNumber(input.Model.Number)
		if err != nil {
			return c.Failure(err)
		}

		_, err = c.Services().Posts.AddComment(post, input.Model.Content)
		if err != nil {
			return c.Failure(err)
		}

		c.Enqueue(tasks.NotifyAboutNewComment(post, input.Model))

		return c.Ok(web.Map{})
	}
}

// UpdateComment changes an existing comment with new content
func UpdateComment() web.HandlerFunc {
	return func(c web.Context) error {
		input := new(actions.EditComment)
		if result := c.BindTo(input); !result.Ok {
			return c.HandleValidation(result)
		}

		err := c.Services().Posts.UpdateComment(input.Model.ID, input.Model.Content)
		if err != nil {
			return c.Failure(err)
		}

		return c.Ok(web.Map{})
	}
}

// AddVote adds current user to given post list of votes
func AddVote() web.HandlerFunc {
	return func(c web.Context) error {
		return addOrRemove(c, c.Services().Posts.AddVote)
	}
}

// RemoveVote removes current user from given post list of votes
func RemoveVote() web.HandlerFunc {
	return func(c web.Context) error {
		return addOrRemove(c, c.Services().Posts.RemoveVote)
	}
}

// Subscribe adds current user to list of subscribers of given post
func Subscribe() web.HandlerFunc {
	return func(c web.Context) error {
		return addOrRemove(c, c.Services().Posts.AddSubscriber)
	}
}

// Unsubscribe removes current user from list of subscribers of given post
func Unsubscribe() web.HandlerFunc {
	return func(c web.Context) error {
		return addOrRemove(c, c.Services().Posts.RemoveSubscriber)
	}
}

func addOrRemove(c web.Context, addOrRemove func(post *models.Post, user *models.User) error) error {
	number, err := c.ParamAsInt("number")
	if err != nil {
		return c.Failure(err)
	}

	post, err := c.Services().Posts.GetByNumber(number)
	if err != nil {
		return c.Failure(err)
	}

	err = addOrRemove(post, c.User())
	if err != nil {
		return c.Failure(err)
	}

	return c.Ok(web.Map{})
}
