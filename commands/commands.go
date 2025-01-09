package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	definederrors "github.com/sohWenMing/aggregator/defined_errors"
	errorutils "github.com/sohWenMing/aggregator/error_utils"
	"github.com/sohWenMing/aggregator/internal/database"
	"github.com/sohWenMing/aggregator/rss_parsing"
)

type handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)

// ############# command struct, used to house all the configured commands with relevant methods ######### //
type commands struct {
	commandMap map[string]handler
}

func (c *commands) ExecCommand(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	handler, ok := c.commandMap[cmd.name]
	if !ok {
		return definederrors.ErrorHandlerNotExist
	}
	handlerErr := handler(cmd, w, state)
	if handlerErr != nil {
		return handlerErr
	}
	return nil
}

func (c *commands) registerAllHandlers() (err error) {

	for _, nameToHandler := range initAllNameToHandlers() {
		err := c.registerHandler(nameToHandler.name, nameToHandler.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) registerAllHandlersTest(nameToHandlers []nameToHandler) (err error) {
	for _, nameToHandler := range nameToHandlers {
		err := c.registerHandler(nameToHandler.name, nameToHandler.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) registerHandler(name string, handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)) (err error) {
	if c.commandMap == nil {
		return fmt.Errorf("pointer to commandMap is nil pointer %w", definederrors.ErrorNilPointer)
	}
	_, found := c.commandMap[name]
	if found {
		return fmt.Errorf("handler %s already exists in commandMap", name)
	}
	c.commandMap[name] = handler
	return nil
}

type nameToHandler struct {
	name    string
	handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)
}

func initAllNameToHandlers() []nameToHandler {
	returnedNameToHandlers := []nameToHandler{
		{"login", handlerLogin},
		{"register", handlerRegisterUser},
		{"reset", handlerResetDatabase},
		{"users", handlerGetUsers},
		{"agg", handlerAgg},
		{"addfeed", handlerAddFeed},
		{"feeds", handlerGetFeeds},
		{"follow", handlerAddFeedFollow},
		{"following", handlerGetFeedFollowsForUser},
	}
	return returnedNameToHandlers

}

func handlerGetUsers(_ enteredCommand, w io.Writer, state *database.State) (err error) {

	users, getUsersErr := state.Db.GetUsers(context.Background())
	if getUsersErr != nil {
		isPqErr, pqErr, rawErr := errorutils.UnwrapPqErr(getUsersErr)
		switch isPqErr {
		case true:
			fmt.Fprintf(w, "error code: %s\n", string(pqErr.Code))
			return fmt.Errorf("postgres error occured: %w", definederrors.ErrorDatabaseErr)
		case false:
			fmt.Fprintln(w, rawErr.Error())
			return rawErr
		}
	}
	for _, user := range users {
		stringBytes := []byte("*" + " " + user.Name)
		if user.Name == state.Cfg.CurrentUserName {
			stringBytes = append(stringBytes, []byte(" (current)")...)
		}
		stringToPrint := string(stringBytes)
		fmt.Fprintln(w, stringToPrint)
	}
	return nil
}

func handlerLogin(cmd enteredCommand, w io.Writer, state *database.State) (retrieveErr error) {
	if len(cmd.args) != 1 {
		return fmt.Errorf("args passed into handlerLogin %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	_, err := state.Db.RetrieveUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Fprintf(w, "user %s could not be retrieved, user is not logged in\n", cmd.args[0])
		return fmt.Errorf("user %s could not be retrieved, user is not logged in %w", cmd.args[0], definederrors.ErrorUserNotFound)
	}

	state.Cfg.SetUser(cmd.args[0], w)
	fmt.Fprintf(w, "user %s is now logged in\n", cmd.args[0])
	return nil
}

func handlerRegisterUser(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 1 {
		return fmt.Errorf("args passed into handlerCreateUser %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	_, createErr := state.Db.CreateUser(context.Background(), params)
	if createErr != nil {
		isPQErr, pqErr, rawErr := errorutils.UnwrapPqErr(createErr)
		if isPQErr {
			if pqErr.Code == "23505" {
				fmt.Fprintf(w, "User %s already exists in database\n", cmd.args[0])
				return fmt.Errorf("user %s already exists %w", cmd.args[0], definederrors.ErrorUserAlreadyExists)
			}
		}
		return rawErr
	}
	fmt.Fprintf(w, "user %s has been added\n", cmd.args[0])
	state.Cfg.SetUser(cmd.args[0], w)
	return nil
}

func handlerResetDatabase(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	state.Db.ResetUsers(context.Background())
	return nil
}

func handlerAgg(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	feedUrl := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(feedUrl, state)
	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Fprintf(w, "the operation timed out")
		}
		return err
	}
	fmt.Fprintf(w, "%v", *feed)
	return nil
}

func handlerTest(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	for _, arg := range cmd.args {
		fmt.Fprintln(w, arg)
	}
	return nil
}

func handlerAddFeed(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 2 {
		return fmt.Errorf("wrong num args passed into handlerAddFeed %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	feedName := cmd.args[0]
	feedUrl := cmd.args[1]
	_, fetchFeedErr := fetchFeed(feedUrl, state)
	if fetchFeedErr != nil {
		switch errorutils.CheckErrTypeMatch(fetchFeedErr, context.DeadlineExceeded) {
		case true:
			fmt.Fprintf(w, "The request to %s timed out", feedUrl)
			return fetchFeedErr
		case false:
			fmt.Fprint(w, fetchFeedErr.Error())
			return fetchFeedErr
		}
	}
	loggedInUser, err := state.Db.RetrieveUser(context.Background(), state.Cfg.CurrentUserName)
	if err != nil {
		fmt.Fprintf(w, "user %s not found in database\n", state.Cfg.CurrentUserName)
		return fmt.Errorf("user %s not found in database %w", state.Cfg.CurrentUserName, definederrors.ErrorUserNotFound)
	}
	rssFeed, err := state.Db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      feedName,
			Url:       feedUrl,
			UserID:    loggedInUser.ID,
		})
	if err != nil {
		fmt.Fprint(w, "error occured when trying to add RssFeed to database")
		return err
	}
	fmt.Fprintf(w, "rss feed values: %v", rssFeed)

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    loggedInUser.ID,
		FeedID:    rssFeed.ID,
	}

	_, feedFollowErr := state.Db.CreateFeedFollow(context.Background(), params)
	if feedFollowErr != nil {
		fmt.Fprint(w, "error occured when attempting to create feedFollow")
		return feedFollowErr
	}
	return nil

}

func handlerGetFeeds(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 0 {
		return definederrors.ErrorWrongNumArgs
	}
	feeds, err := state.Db.GetFeeds(context.Background())
	if err != nil {
		isPqErr, pqErr, rawErr := errorutils.UnwrapPqErr(err)
		if isPqErr {
			fmt.Fprintf(w, "pqerr error code: %s", string(pqErr.Code))
			return pqErr
		}
		fmt.Fprintln(w, "error occured while getting feeds")
		return rawErr
	}
	for _, feed := range feeds {
		feedInfo := fmt.Sprintf("feed name: %s feed url: %s user name: %s\n", feed.Feedname, feed.Feedurl, feed.Username)
		fmt.Fprint(w, feedInfo)
	}
	return nil
}

func handlerAddFeedFollow(cmd enteredCommand, w io.Writer, state *database.State) (err error) {

	if len(cmd.args) != 1 {
		fmt.Fprint(w, definederrors.ErrorWrongNumArgs.Error())
		return definederrors.ErrorWrongNumArgs
	}

	//first, need to get the user id from the db
	currUser, err := retrieveUserFromDB(state, w)
	if err != nil {
		return err
	}
	currUserId := currUser.ID
	feedId, err := state.Db.GetFeedIdByURL(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Fprintf(w, "feed with url %s could not be found", cmd.args[0])
		return err
	}
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currUserId,
		FeedID:    feedId,
	}

	feedFollowRow, err := state.Db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		fmt.Fprint(w, "error occured when attempting to create feedFollow")
		return err
	}
	fmt.Fprintf(w, "%s is now following %s\n", feedFollowRow.UserName, feedFollowRow.FeedName)
	return nil

}

func handlerGetFeedFollowsForUser(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 0 {
		fmt.Fprint(w, definederrors.ErrorWrongNumArgs.Error())
		return err
	}
	currUser, err := retrieveUserFromDB(state, w)
	if err != nil {
		return err
	}
	feeds, err := state.Db.GetFeedFollowForUser(context.Background(), currUser.ID)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return err
	}
	fmt.Fprintf(w, "feeds followed by user %s\n", currUser.Name)
	for _, feed := range feeds {
		fmt.Fprintf(w, "* %s - %s\n", feed.FeedName, feed.FeedUrl)
	}
	return nil
}

func fetchFeed(feedURL string, state *database.State) (feed *rss_parsing.RSSFeed, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "gator")
	res, err := state.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	rssFeed, err := rss_parsing.ParseRSS(resBody)
	if err != nil {
		return nil, err
	}
	return &rssFeed, nil

}

func retrieveUserFromDB(state *database.State, w io.Writer) (database.User, error) {
	currUser, err := state.Db.RetrieveUser(context.Background(), state.Cfg.CurrentUserName)
	if err != nil {
		fmt.Fprintf(w, "user %s could not be retrieved from database", state.Cfg.CurrentUserName)
		return database.User{}, err
	}
	return currUser, nil
}

// called at the main program, used to initialise the commandMap so that it can be written to
func InitCommands() (commandsPtr *commands) {
	returnedCommands := commands{}
	commandMap := make(map[string]handler)
	returnedCommands.commandMap = commandMap
	returnedCommands.registerAllHandlers()
	return &returnedCommands
}

// function to parse the input from os.Args, if no error should return a parsed enteredCommand
func ParseCommand(args []string) (cmd enteredCommand, err error) {
	returnedCmd := enteredCommand{}
	switch len(args) {
	case 0:
		return returnedCmd,
			fmt.Errorf("no arguments passed into ParseCommand %w",
				definederrors.ErrorNoArgs)
	case 1:
		if args[0] != "agg" {
			return returnedCmd,
				fmt.Errorf("only one arguement passed into ParseCommand arg:%s %w",
					args[0], definederrors.ErrorWrongNumArgs)
		}
		returnedCmd.name = strings.ToLower(args[1])
		returnedCmd.args = []string{}
		return returnedCmd, nil
	default:
		returnedCmd.name = strings.ToLower(args[1])
		returnedCmd.args = args[2:]
		return returnedCmd, nil
	}
}

type enteredCommand struct {
	name string
	args []string
}

/*
what needs to be achieved

when something is entered as a command, it needs to be parsed into an entered command
the entered command's name has to be used to correlate to the map of strings to handlers


*/
