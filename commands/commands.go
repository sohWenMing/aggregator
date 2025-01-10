package commands

import (
	"context"
	"database/sql"
	"errors"
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
		{"addfeed", middleWareLoggedIn(handlerAddFeed)},
		{"feeds", handlerGetFeeds},
		{"follow", middleWareLoggedIn(handlerAddFeedFollow)},
		{"following", middleWareLoggedIn(handlerGetFeedFollowsForUser)},
		{"unfollow", middleWareLoggedIn(handlerRemoveFeedFollow)},
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
	loggedInUser, err := state.Db.RetrieveUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Fprintf(w, "user %s could not be retrieved, user is not logged in\n", cmd.args[0])
		return fmt.Errorf("user %s could not be retrieved, user is not logged in %w", cmd.args[0], definederrors.ErrorUserNotFound)
	}

	state.Cfg.SetUser(cmd.args[0], w)
	state.Cfg.CurrentUser.ID = loggedInUser.ID
	state.Cfg.CurrentUser.Name = loggedInUser.Name
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
	createdUser, createErr := state.Db.CreateUser(context.Background(), params)

	if createErr != nil {
		isPQErr, isUniqueViolation, _, rawErr := database.CheckPqErr(createErr)

		if isUniqueViolation {
			fmt.Fprintf(w, "User %s already exists in database\n", cmd.args[0])
			return fmt.Errorf("user %s already exists %w", cmd.args[0], definederrors.ErrorUserAlreadyExists)
		}

		if isPQErr {
			fmt.Fprintln(w, definederrors.ErrorDatabaseErr.Error())
			return definederrors.ErrorDatabaseErr
		}
		return rawErr
	}

	fmt.Fprintf(w, "user %s has been added\n", cmd.args[0])
	state.Cfg.SetUser(cmd.args[0], w)
	state.Cfg.CurrentUser.ID = createdUser.ID
	state.Cfg.CurrentUser.Name = createdUser.Name
	return nil
}

func handlerResetDatabase(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	state.Db.ResetUsers(context.Background())
	return nil
}

func handlerAgg(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 1 {
		fmt.Fprint(w, definederrors.ErrorWrongNumArgs.Error())
		return definederrors.ErrorWrongNumArgs
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		fmt.Fprint(w, err.Error())
		return err
	}

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeErr := scrapeFeeds(w, state)
		if scrapeErr != nil {
			fmt.Fprintln(w, scrapeErr.Error())
		}
	}
}

func scrapeFeeds(w io.Writer, state *database.State) (err error) {
	feedToFetch, err := state.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println("error happened at feedToFetch")
		return err
	}
	params := database.MarkFetchedFeedParams{
		UpdatedAt: time.Now(),
		ID:        feedToFetch.ID,
	}
	markErr := state.Db.MarkFetchedFeed(context.Background(), params)
	if markErr != nil {
		return markErr
	}

	feed, err := fetchFeed(feedToFetch.Url, state)
	if err != nil {
		if err == context.DeadlineExceeded {
			return errors.New("the operation timed out")
		}
	}
	fmt.Fprintf(w, "Feed Name: %s\n", feed.Channel.Title)
	items := feed.Channel.RSSItems
	for _, item := range items {
		writeItemToDB(feedToFetch, w, item, state)
	}
	return nil
}

func writeItemToDB(feed database.Feed, w io.Writer, item rss_parsing.RSSItem, state *database.State) {

	params := database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       item.Title,
		Url:         item.Link,
		Description: genNullableString(item.Description),
		PublishedAt: parseDate(item.PubDate),
		FeedID:      feed.ID,
	}

	_, err := state.Db.CreatePost(context.Background(), params)
	if err != nil {
		isPQErr, isUniqueViolation, pqErr, rawErr := database.CheckPqErr(err)

		if isUniqueViolation {
			return
		}

		if isPQErr {
			fmt.Fprintf(w, "PqErrCode: %s\n", pqErr.Code)
			fmt.Fprintf(w, "PqErrMsg: %s\n", pqErr.Message)
			return
		}
		fmt.Fprintln(w, rawErr.Error())
		return
	}

}

func genNullableString(input string) (sqlNullableString sql.NullString) {

	if len(input) == 0 {
		sqlNullableString.Valid = false
		return sqlNullableString
	}
	sqlNullableString.String = input
	sqlNullableString.Valid = true
	return sqlNullableString

}

func parseDate(timestamp string) (sqlNullableTime sql.NullTime) {
	parsedDate, err := time.Parse(time.RFC1123Z, timestamp)
	var nullTime sql.NullTime
	if err != nil {
		nullTime.Valid = false
		return nullTime
	}
	nullTime.Time = parsedDate
	nullTime.Valid = true
	return nullTime
}

func handlerTest(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	for _, arg := range cmd.args {
		fmt.Fprintln(w, arg)
	}
	return nil
}

func middleWareLoggedIn(enteredHandler handler) (returnedHandler handler) {
	return func(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
		loggedInUser, err := state.Db.RetrieveUser(context.Background(), state.Cfg.CurrentUserName)
		if err != nil {
			fmt.Printf("user with username %s could not be found in database", state.Cfg.CurrentUserName)
			return err
		}
		state.Cfg.CurrentUser.ID = loggedInUser.ID
		state.Cfg.CurrentUser.Name = loggedInUser.Name
		return enteredHandler(cmd, w, state)
	}
}

func handlerRemoveFeedFollow(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong num args passed into handlerRemoveFeedFollow, %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	params := database.DeleteFeedFollowParams{
		UserID: state.Cfg.CurrentUser.ID,
		Url:    cmd.args[0],
	}
	deleteErr := state.Db.DeleteFeedFollow(context.Background(), params)
	if deleteErr != nil {
		return fmt.Errorf("error occured when attempting to remove feed follow")
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
	// loggedInUser, err := state.Db.RetrieveUser(context.Background(), state.Cfg.CurrentUserName)
	// if err != nil {
	// 	fmt.Fprintf(w, "user %s not found in database\n", state.Cfg.CurrentUserName)
	// 	return fmt.Errorf("user %s not found in database %w", state.Cfg.CurrentUserName, definederrors.ErrorUserNotFound)
	// }
	rssFeed, err := state.Db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      feedName,
			Url:       feedUrl,
			UserID:    state.Cfg.CurrentUser.ID,
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
		UserID:    state.Cfg.CurrentUser.ID,
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

	feedId, err := state.Db.GetFeedIdByURL(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Fprintf(w, "feed with url %s could not be found", cmd.args[0])
		return err
	}
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    state.Cfg.CurrentUser.ID,
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

	feeds, err := state.Db.GetFeedFollowForUser(context.Background(), state.Cfg.CurrentUser.ID)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return err
	}
	fmt.Fprintf(w, "feeds followed by user %s\n", state.Cfg.CurrentUserName)
	for _, feed := range feeds {
		fmt.Fprintf(w, "* %s - %s\n", feed.FeedName, feed.FeedUrl)
	}
	return nil
}

func fetchFeed(feedURL string, state *database.State) (feed *rss_parsing.RSSFeed, err error) {

	req, cancel, err := getReqAndCancelFunc(feedURL)
	if err != nil {
		return nil, err
	}
	defer cancel()
	return makeRSSReq(*state.Client, req)
}

func testFetchFeed(feedURL string) (feed *rss_parsing.RSSFeed, err error) {
	req, cancel, err := getReqAndCancelFunc(feedURL)
	if err != nil {
		return nil, err
	}
	defer cancel()
	return makeRSSReq(http.Client{}, req)
}

func makeRSSReq(client http.Client, req *http.Request) (*rss_parsing.RSSFeed, error) {
	res, err := client.Do(req)
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

func getReqAndCancelFunc(feedURL string) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	req.Header.Set("user-agent", "gator")
	return req, cancel, nil
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
		return returnedCmd,
			fmt.Errorf("only one arguement passed into ParseCommand arg:%s %w",
				args[0], definederrors.ErrorWrongNumArgs)

	default:
		returnedCmd.name = strings.ToLower(args[1])
		if len(args) == 2 {
			return returnedCmd, nil
		}
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
