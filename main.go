package main

import (
    "database/sql"
    "errors"
    "fmt"
    "gioui.org/app"
    "gioui.org/font"
    "gioui.org/io/system"
    "gioui.org/layout"
    "gioui.org/op"
    "gioui.org/text"
    "gioui.org/unit"
    "gioui.org/widget"
    "gioui.org/widget/material"
    "github.com/casbin/casbin/v2"
    "github.com/casbin/casbin/v2/model"
    "github.com/casbin/casbin/v2/persist"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "image/color"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)


func main() {
    // Create sqlite3 object to fetch data from DB
    s3db := openDb()

    // The app runs in a go routine
    go func() {
        // Define a window instance - we could create multiple windows if needed
        signInWindow := new(app.Window)
        err          := runSignIn(signInWindow, s3db)

        if err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()

    // Gio needs this line to hand over the routine to main program
    app.Main()

    // defer close the connection to sqlite
    defer s3db.Close()
}


// Functions for handling windows
func runSignIn(inWindow *app.Window, inS3db *sql.DB) error {
    var ops                 op.Ops 			  // List of operations gio library uses to know what needs to be shown in a window
    var signInBtn           widget.Clickable
    var usernameTextbox     widget.Editor
    var passwordTextbox     widget.Editor
    var errorMsg            string

    var theme  = material.NewTheme()

    titleText := "Very Simple-teab app"
    btnText   := "Sign In"

    for {
        event := inWindow.Event()

        switch eventType := event.(type) {
        // This one triggers when the window is closed
        case app.DestroyEvent:
            return eventType.Err
        // FrameEvent runs before the window is presented on screen
        case app.FrameEvent:
            // This layout context is used for managing the rendering state of the window
            gtx      := app.NewContext(&ops, eventType)

            // Set an action for button click
            if signInBtn.Clicked(gtx) {
                var username string
                var password string

                username = usernameTextbox.Text()
                password = passwordTextbox.Text()

                if len(username) > 0 && len(password) > 0 {
                    // Check sign in credentials
                    success, userID := checkSignIn(username, password, inS3db)

                    if !success {
                        fmt.Printf("Sign-in failed: %s // %s\n", username, password)
                        break
                    }

                    // Open main window and close sign in
                    go func() {
                        mainWindow := new(app.Window)
                        inWindow.Perform(system.ActionMinimize)
                        err        := runApp(mainWindow, userID, username, inS3db)

                        if err != nil {
                            log.Fatal(err)
                        }

                        defer inWindow.Perform(system.ActionClose)
                    }()
                } else {
                    errorMsg   = "Please enter a username and a password"
                }
            }

            layout.Flex{
                // Vertical alignment, from top to bottom
                Axis: layout.Vertical,
                // Empty space is left at the start, i.e. at the top
                Spacing: layout.SpaceStart,
            }.Layout(gtx,
                // Title on top - in Flex Layout Flexed objects start filling from the top
                layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
                    maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
                    return titleElement(gtx, theme, titleText, 1, maroon)
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(30)}.Layout),

                // Error box, if there is an error to show
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return errorBoxElement(gtx, theme, errorMsg)
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(50)}.Layout),

                // Textbox for username
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return inputBoxElement(gtx, theme, &usernameTextbox, "Enter username")
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

                // Textbox for password
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    // Hide user's input with a mask
                    passwordTextbox.Mask = '•'
                    return inputBoxElement(gtx, theme, &passwordTextbox, "Enter password")
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),

                // Button for submitting username and password
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return btnElement(gtx, theme, &signInBtn, btnText)
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),
            )

            // Pass the drawing operations to the GPU
            eventType.Frame(gtx.Ops)
        }
    }
}


func runApp(inWindow *app.Window, inUserID int, inUsername string, inS3db *sql.DB) error {
    var ops                 op.Ops 			  // List of operations gio library uses to know what needs to be shown in a window
    var inputConfirmBtn     widget.Clickable
    var clientTextbox       widget.Editor
    var timeTextbox         widget.Editor
    var clickCntText        string

    var theme               = material.NewTheme()

    titleText               := "Very Simple showcase app with unnecessarily long title"
    subTitleText            := fmt.Sprintf("Welcome back %s! We did not miss you!", inUsername)
    adminTextAllowed        := fmt.Sprintf("Your user ID is %d, probably", inUserID)
    adminTextDenied         := fmt.Sprintf("Only Admin users can view their ID, you are just a minion")
    btnText                 := "Confirm"
    clicksCnt               := 0

    // Init Casbin
    userEnforcer := initCasbinEnforcers()

    for {
        event := inWindow.Event()

        switch eventType := event.(type) {
        // This one triggers when the window is closed
        case app.DestroyEvent:
            return eventType.Err
        // FrameEvent runs before the window is presented on screen
        case app.FrameEvent:
            // This layout context is used for managing the rendering state of the window
            gtx      := app.NewContext(&ops, eventType)

            // Set an action for button click
            if inputConfirmBtn.Clicked(gtx) && len(clientTextbox.Text()) > 0 && len(timeTextbox.Text()) > 0 {
                canReportClientName := enforceCasbin(userEnforcer, fmt.Sprintf("u%d", inUserID), "inputbox_client_name", "write")
                canReportTimeSpent  := enforceCasbin(userEnforcer, fmt.Sprintf("u%d", inUserID), "inputbox_time_spent",  "write")

                if canReportClientName && canReportTimeSpent {
                    // Increase on click
                    clicksCnt += 1
                    clientTextbox.SetText("")
                    timeTextbox.SetText("")
                    clickCntText = fmt.Sprintf("Confirmed the report text: %d times", clicksCnt)
                } else {
                    clientTextbox.SetText("")
                    timeTextbox.SetText("")
                    clickCntText = "You shall not pass!.. the reports"
                }
            }

            layout.Flex{
                // Vertical alignment, from top to bottom
                Axis: layout.Vertical,
                // Empty space is left at the start, i.e. at the top
                Spacing: layout.SpaceStart,
            }.Layout(gtx,
                // Title on top - in Flex Layout Flexed objects start filling from the top
                layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
                    maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
                    return titleElement(gtx, theme, titleText, 1, maroon)
                }),

                // Empty spacer
                layout.Flexed(1,layout.Spacer{Height: unit.Dp(10)}.Layout),

                // Subtitle
                layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
                    subColor := color.NRGBA{R: 12, G: 13, B: 114, A: 240}
                    return titleElement(gtx, theme, subTitleText, 2, subColor)
                }),

                // Empty spacer
                layout.Flexed(1,layout.Spacer{Height: unit.Dp(20)}.Layout),

                // Admin text
                layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
                    var adminText        string
                    var canViewAdminText bool
                    //  Casbin check
                    canViewAdminText = enforceCasbin(userEnforcer, fmt.Sprintf("u%d", inUserID), "admin_text", "read")

                    if canViewAdminText {
                        adminText = adminTextAllowed
                    } else {
                        adminText = adminTextDenied
                    }
                    newColor := color.NRGBA{R: 127, G: 152, B: 42, A: 250}

                    return reportBoxElement(gtx, theme, adminText, newColor)
                }),

                // Ticks and clicks count textbox
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    // Casbin check report text
                    canViewReportText := enforceCasbin(userEnforcer, fmt.Sprintf("u%d", inUserID), "report_text", "read")
                    someColor         := color.NRGBA{R: 127, G: 152, B: 0, A: 160}

                    if canViewReportText {
                        return reportBoxElement(gtx, theme, clickCntText, someColor)
                    } else {
                        return layout.Dimensions{}
                    }
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(30)}.Layout),

                // Input box
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return inputBoxElement(gtx, theme, &clientTextbox, "Input for T&B client name")
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

                // Input box
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return inputBoxElement(gtx, theme, &timeTextbox, "Input for T&B time spent")
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(50)}.Layout),

                // Button for counting clicks
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return btnElement(gtx, theme, &inputConfirmBtn, btnText)
                }),

                // Empty spacer
                layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),
            )

            // Pass the drawing operations to the GPU
            eventType.Frame(gtx.Ops)
        }
    }
}


// Functions for specific elements shown in windows
func errorBoxElement(inGTX layout.Context, inTheme *material.Theme, inErrTxt string) layout.Dimensions {
    // Define a large label with a text
    title          := material.H4(inTheme, inErrTxt)
    // Change the color of the label
    red            := color.NRGBA{R: 200, G: 0, B: 0, A: 192}
    title.Color     = red
    // Change the alignment position of the label
    title.Alignment = text.Middle
    // Draw the label to the layout context
    return title.Layout(inGTX)
}


func titleElement(inGTX layout.Context, inTheme *material.Theme, inTxt string, inSize int, inColor color.NRGBA) layout.Dimensions {
    // Define a large label with a text
    var title material.LabelStyle

    if inSize == 1 {
        title          = material.H3(inTheme, inTxt)
    } else {
        title          = material.H4(inTheme, inTxt)
    }
    // Change the color of the label
    title.Color     = inColor
    // Change the alignment position of the label
    title.Alignment = text.Middle
    // Draw the label to the layout context
    return title.Layout(inGTX)
}


func reportBoxElement(inGTX layout.Context, inTheme *material.Theme, inClickCntText string, inColor color.NRGBA) layout.Dimensions {
    // Define a large label with a text
    changingTxt          := material.Label(inTheme, unit.Sp(inGTX.Sp(12)), inClickCntText)
    // Change the color of the label
    changingTxt.Color     = inColor
    // Change the alignment position of the label
    changingTxt.Alignment = text.Middle
    // Draw the label to the layout context
    return changingTxt.Layout(inGTX)
}


func inputBoxElement(inGTX layout.Context, inTheme *material.Theme, inInputTextbox *widget.Editor, inHintText string) layout.Dimensions {
    // This part keeps it in the center without stretching the widget to the window's edge
    return layout.Flex{
        Axis:    layout.Horizontal,
        Spacing: layout.SpaceAround,
    }.Layout(inGTX,
        // Empty flexible space to push content to center
        layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
            return layout.Dimensions{}
        }),
        layout.Rigid(func(gtx layout.Context) layout.Dimensions {
            // Set fixed width
            gtx.Constraints.Min.X = gtx.Dp(300)
            gtx.Constraints.Max.X = gtx.Dp(300)
            // Define an editable input box
            inputBox             := material.Editor(inTheme, inInputTextbox, inHintText)
            // Change the font of a box's text
            inputBox.Font         = font.Font{Typeface: "Light"}
            // Put the box in the middle of the window
            inInputTextbox.Alignment = text.Middle
            // Add a border so the box actually has a form
            border := widget.Border{
                Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
                CornerRadius: unit.Dp(3),
                Width:        unit.Dp(2),
            }
            // Draw the input box to the layout context, but put it in the center of the window
            return border.Layout(gtx, inputBox.Layout)
        }),
        // Empty flexible space to balance layout
        layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
            return layout.Dimensions{}
        }),
    )
}


func btnElement(inGTX layout.Context, inTheme *material.Theme, inInputConfirmBtn *widget.Clickable, inBtnText string) layout.Dimensions {
    return layout.Flex{
        Axis:    layout.Horizontal,
        Spacing: layout.SpaceAround,
    }.Layout(inGTX,
        // Empty flexible space to push content to center
        layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
            return layout.Dimensions{}
        }),
        // Actual button with fixed width
        layout.Rigid(func(gtx layout.Context) layout.Dimensions {
            // Set fixed width
            gtx.Constraints.Min.X   = gtx.Dp(150)
            gtx.Constraints.Max.X   = gtx.Dp(150)
            // Define a button with a text
            btn                    := material.Button(inTheme, inInputConfirmBtn, inBtnText)
            // Change the font of a button's text
            btn.Font                = font.Font{Typeface: "ExtraLight"}
            btn.Inset               = layout.UniformInset(10)
            // Draw the button to the layout context in the center of the window
            return btn.Layout(gtx)
        }),
        // Empty flexible space to balance layout
        layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
            return layout.Dimensions{}
        }),
    )
}


// DB functions
func openDb () *sql.DB {

    db, err := sql.Open("sqlite3", "data/database/showcase_db")

    if err != nil {
        log.Fatal(err)
    }

    // Test the connection
    err = db.Ping()

    if err != nil {
        log.Fatal(err)
    }

    return db
}


func executeQuery(query string, db *sql.DB) bool {
    _, err := db.Exec(query)

    if err != nil {
        log.Fatal(err)
        return false
    }

    return true
}


func runReturnQuery(query string, db *sql.DB) *sql.Rows {
    results, err := db.Query(query)

    if err != nil {
        log.Fatal(err)
    }

    return results
}


func checkSignIn(inUsername string, inPassword string, inDB *sql.DB) (bool, int) {
    var userId int

    signInQuery := fmt.Sprintf(`
SELECT
    user_id
FROM
    user_dim
WHERE
        username = '%s'
    AND password = '%s'
	`, inUsername, inPassword)

    results := runReturnQuery(signInQuery, inDB)

    results.Next()
    scanErr := results.Scan(&userId)
    if scanErr != nil {
        log.Print(scanErr)
        return false, 0
    }
    fmt.Printf("UserId = %d\n", userId)

    return true, userId
}

//Casbin functions
func initCasbinEnforcers() *casbin.Enforcer {
    // Create connection to DB for Gorm
    casbinDB, dbOpenErr := gorm.Open(sqlite.Open("data/database/showcase_db"), &gorm.Config{})
    if dbOpenErr        != nil {
       log.Fatalf("Failed to connect to database for Casbin: %v", dbOpenErr)
    }
    db, _ := casbinDB.DB()
    defer db.Close()

    userAdapter, userAdapterErr := NewCustomAdapter("data/database/showcase_db")
    if userAdapterErr           != nil {
      log.Fatalf("Failed to create userAdapter: %v", userAdapterErr)
    }

    // Load Casbin userEnforcer
    userEnforcer, userEnforcerErr := casbin.NewEnforcer("data/steaby_casbin_model.conf", userAdapter)
    if userEnforcerErr            != nil {
        log.Fatalf("Failed to create user enforcer: %v", userEnforcerErr)
    }

    // Load user policies from DB
    userPoliciesErr    := userEnforcer.LoadPolicy()
    if userPoliciesErr != nil {
        log.Fatalf("Failed to load user policy: %v", userPoliciesErr)
    }

    return userEnforcer
}

func enforceCasbin(inEnforcer *casbin.Enforcer, subject string, object string, action string) bool {
    ok, enfErr := inEnforcer.Enforce(subject, object, action)
    if enfErr != nil {
        log.Fatalf("Failed to check the policy: %v", enfErr)
    }
    fmt.Printf("ok? = %v\n", ok)

    return ok
}


// <editor-fold desc="CustomAdapter">

// CustomAdapter Define structure and functions for custom policy adapter
type CustomAdapter struct {
    db *sql.DB
}

func NewCustomAdapter(dbPath string) (*CustomAdapter, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    return &CustomAdapter{db: db}, nil
}

func (a *CustomAdapter) Close() error {
    return a.db.Close()
}

// LoadPolicy loads all policies from the database into Casbin
func (a *CustomAdapter) LoadPolicy(model model.Model) error {
    rows, err := a.db.Query("SELECT subject, object, action, effect FROM casbin_rule")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var line, sub, obj, act, eff    string
        var nil1, nil2                  interface{}

        isUserPolicy := true

        if err := rows.Scan(&sub, &obj, &act, &eff); err != nil {
            // If error was encountered it means that action and effect are missing -> which means that we are trying to load a role policy
            isUserPolicy = false

            s2err := rows.Scan(&sub, &obj, &nil1, &nil2)
            if s2err != nil {
                return s2err
            }
        }

        if isUserPolicy {
            line = fmt.Sprintf("p, %s, %s, %s, %s", sub, obj, act, eff) // Casbin policy format
        } else {
            line = fmt.Sprintf("g, %s, %s", sub, obj)
        }

        persist.LoadPolicyLine(line, model)
    }

    return nil
}

// SavePolicy saves all policies
// TODO will need improvement
func (a *CustomAdapter) SavePolicy(model model.Model) error {
    _, err := a.db.Exec("DELETE FROM casbin_rule") // Clear existing policies
    if err != nil {
        return err
    }

    stmt, err := a.db.Prepare("INSERT INTO casbin_rule (subject, object, action, effect) VALUES (?, ?, ?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, assertion := range model["p"]["p"].Policy {
        _, err := stmt.Exec(assertion[0], assertion[1], assertion[2], assertion[3])
        if err != nil {
            return err
        }
    }

    return nil
}

// AddPolicy inserts a single policy rule
// TODO will need improvement
func (a *CustomAdapter) AddPolicy(sec string, ptype string, rule []string) error {
    _, err := a.db.Exec("INSERT INTO casbin_rule (subject, object, action, effect) VALUES (?, ?, ?, ?)", rule[0], rule[1], rule[2], rule[3])
    return err
}

// RemovePolicy deletes a policy rule
// TODO will need improvement
func (a *CustomAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
    _, err := a.db.Exec("DELETE FROM casbin_rule WHERE subject = ? AND object = ? AND action = ? AND effect = ?", rule[0], rule[1], rule[2], rule[3])
    return err
}

// RemoveFilteredPolicy removes a filtered policy
// TODO will need improvement
func (a *CustomAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
    return errors.New("don't tell my boss, but this is not implemented")
}

//</editor-fold>
