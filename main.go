package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type GameState struct {
	Tiles          []Tile
	MistakesLeft   int
	Date           string
	Message        string
	SelectedTiles  []int
	CorrectGroups  [][]string
	SolvedGroups   [][]string
	CurrentGroupID int
}

type Tile struct {
	ID       int
	Word     string
	Selected bool
	GroupID  int // 0 means not assigned to a group yet
}

func main() {

	gameState := initializeGame()

	// Define custom template functions
	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	}

	// Create template with the function map
	tmpl := template.Must(template.New("game").Funcs(funcMap).Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fun-nections</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/htmx/1.9.6/htmx.min.js"></script>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
        }
       
        body {
            max-width: 600px;
            margin: 0 auto;
            padding: 0 20px;
        }
       
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 20px 0;
            border-bottom: 1px solid #e0e0e0;
            margin-bottom: 20px;
        }
       
        h1 {
            font-size: 24px;
            font-weight: bold;
        }
       
        .date {
            font-weight: normal;
            font-size: 20px;
        }
       
        .icons {
            display: flex;
            gap: 16px;
            padding: 10px 0;
            border-bottom: 1px solid #e0e0e0;
            margin-bottom: 20px;
            justify-content: flex-end;
        }
       
        .icon {
            font-size: 20px;
            cursor: pointer;
        }
       
        .instructions {
            text-align: center;
            margin-bottom: 20px;
            font-size: 16px;
        }
       
        .grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 10px;
            margin-bottom: 20px;
        }
       
        .tile {
            background-color: #f0f0ea;
            border-radius: 5px;
            padding: 20px 10px;
            display: flex;
            justify-content: center;
            align-items: center;
            text-align: center;
            font-weight: bold;
            font-size: 14px;
            cursor: pointer;
            user-select: none;
            height: 70px;
        }
       
        .tile:hover {
            background-color: #e8e8e0;
        }
       
        .tile.selected {
            background-color: #c9d3f8;
        }
       
        .tile.group-1 {
            background-color: #fbd400;
            color: #000;
        }
       
        .tile.group-2 {
            background-color: #68bc36;
            color: #fff;
        }
       
        .tile.group-3 {
            background-color: #3a96dd;
            color: #fff;
        }
       
        .tile.group-4 {
            background-color: #9b59b6;
            color: #fff;
        }
       
        .game-footer {
            text-align: center;
            margin-top: 20px;
        }
       
        .message {
            margin-bottom: 10px;
            min-height: 20px;
            font-weight: bold;
        }
       
        .mistakes {
            margin-bottom: 20px;
            font-size: 14px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 5px;
        }
       
        .mistake-dot {
            display: inline-block;
            width: 12px;
            height: 12px;
            background-color: #666;
            border-radius: 50%;
        }
       
        .buttons {
            display: flex;
            gap: 10px;
            justify-content: center;
        }
       
        .button {
            padding: 10px 20px;
            border-radius: 20px;
            font-size: 14px;
            cursor: pointer;
            border: 1px solid #ccc;
            background-color: white;
        }
       
        .submit-button {
            background-color: white;
        }
    </style>
</head>
<body>
    <header>
        <h1>Fun-nections <span class="date">{{.Date}}</span></h1>
    </header>
   
    <div class="icons">
        <div class="icon">🔨</div>
        <div class="icon">❓</div>
    </div>
   
    <div class="instructions">
        {{.Message}}
    </div>
   
    <div class="grid" id="gameGrid">
        {{range .Tiles}}
            <div class="tile {{if .Selected}}selected{{end}} {{if gt .GroupID 0}}group-{{.GroupID}}{{end}}"
                 hx-post="/select-tile"
                 hx-vals='{"id": "{{.ID}}"}'
                 hx-target="body"
                 hx-swap="innerHTML"
                 {{if gt .GroupID 0}}style="pointer-events: none;"{{end}}>
                {{.Word}}
            </div>
        {{end}}
    </div>
   
    <div class="game-footer">
        <div class="mistakes">
            Mistakes Remaining:
            {{range $i := seq 1 .MistakesLeft}}
                <div class="mistake-dot"></div>
            {{end}}
        </div>
       
        <div class="buttons">
            <button class="button" hx-post="/shuffle" hx-target="body" hx-swap="innerHTML">Shuffle</button>
            <button class="button" hx-post="/deselect-all" hx-target="body" hx-swap="innerHTML">Deselect All</button>
            <button class="button submit-button" hx-post="/submit" hx-target="body" hx-swap="innerHTML">Submit</button>
        </div>
    </div>
    <script>
    function toggleSelect(element) {
      element.classList.toggle('selected');
    }
   
    function shuffle() {
      const grid = document.getElementById('gameGrid');
      for (let i = grid.children.length; i >= 0; i--) {
        grid.appendChild(grid.children[Math.random() * i | 0]);
      }
    }
   
    function deselectAll() {
      const selectedTiles = document.querySelectorAll('.selected');
      selectedTiles.forEach(tile => {
        tile.classList.remove('selected');
      });
    }
   
    function submit() {
      const selectedTiles = document.querySelectorAll('.selected');
      if (selectedTiles.length !== 4) {
        alert('Please select exactly 4 tiles before submitting.');
        return;
      }
      // In a real game, we'd check if the selection is correct here
      // For demo purposes, we'll just deselect all tiles
      deselectAll();
    }
  </script>
</body>
</html>
`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, gameState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/select-tile", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		tileID := 0
		fmt.Sscanf(id, "%d", &tileID)

		// Toggle selection
		for i := range gameState.Tiles {
			if gameState.Tiles[i].ID == tileID {
				gameState.Tiles[i].Selected = !gameState.Tiles[i].Selected
				break
			}
		}

		// Update selected tiles list
		gameState.SelectedTiles = []int{}
		for _, tile := range gameState.Tiles {
			if tile.Selected && len(gameState.SelectedTiles) < 5 {
				gameState.SelectedTiles = append(gameState.SelectedTiles, tile.ID)
			}
		}

		if err := tmpl.Execute(w, gameState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		selectedTiles := []Tile{}
		selectedWords := []string{}

		for _, tile := range gameState.Tiles {
			if tile.Selected {
				selectedTiles = append(selectedTiles, tile)
				selectedWords = append(selectedWords, tile.Word)
			}
		}

		if len(selectedTiles) != 4 {
			gameState.Message = "Please select exactly 4 tiles."
			if err := tmpl.Execute(w, gameState); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Check if the selected tiles form a valid group
		isCorrect := false
		groupIndex := -1

		for i, group := range gameState.CorrectGroups {
			if containsSameElements(selectedWords, group) {
				isCorrect = true
				groupIndex = i
				break
			}
		}

		if isCorrect {
			gameState.CurrentGroupID++
			groupID := gameState.CurrentGroupID

			// Add to solved groups
			gameState.SolvedGroups = append(gameState.SolvedGroups, selectedWords)

			// Mark tiles as solved
			for i := range gameState.Tiles {
				for _, selected := range selectedTiles {
					if gameState.Tiles[i].ID == selected.ID {
						gameState.Tiles[i].Selected = false
						gameState.Tiles[i].GroupID = groupID
					}
				}
			}

			// Remove from available groups
			if groupIndex >= 0 {
				gameState.CorrectGroups = append(gameState.CorrectGroups[:groupIndex], gameState.CorrectGroups[groupIndex+1:]...)
			}

			gameState.Message = "Correct group!"
			gameState.SelectedTiles = []int{}

			if len(gameState.CorrectGroups) == 0 {
				gameState.Message = "You win! All groups found!"
			}
		} else {
			gameState.MistakesLeft--
			gameState.Message = "Incorrect group. Try again."

			// Deselect all tiles
			for i := range gameState.Tiles {
				gameState.Tiles[i].Selected = false
			}
			gameState.SelectedTiles = []int{}

			if gameState.MistakesLeft <= 0 {
				gameState.Message = "Game over! No more mistakes allowed."
			}
		}

		if err := tmpl.Execute(w, gameState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/shuffle", func(w http.ResponseWriter, r *http.Request) {
		// Only shuffle tiles that aren't part of a solved group
		unsolvedTiles := []Tile{}
		solvedTiles := []Tile{}

		for _, tile := range gameState.Tiles {
			if tile.GroupID == 0 {
				unsolvedTiles = append(unsolvedTiles, tile)
			} else {
				solvedTiles = append(solvedTiles, tile)
			}
		}

		// Shuffle unsolved tiles
		rand.Shuffle(len(unsolvedTiles), func(i, j int) {
			unsolvedTiles[i], unsolvedTiles[j] = unsolvedTiles[j], unsolvedTiles[i]
		})

		// Combine back together
		gameState.Tiles = append(unsolvedTiles, solvedTiles...)

		if err := tmpl.Execute(w, gameState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/deselect-all", func(w http.ResponseWriter, r *http.Request) {
		for i := range gameState.Tiles {
			gameState.Tiles[i].Selected = false
		}
		gameState.SelectedTiles = []int{}

		if err := tmpl.Execute(w, gameState); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initializeGame() *GameState {
	correctGroups := [][]string{
		{"A", "B", "C", "D"},
		{"E", "F", "G", "H"},
		{"I", "J", "K", "L"},
		{"M", "N", "O", "N"},
	}

	// Create tiles
	allWords := []string{
		"A", "B", "C", "D",
		"E", "F", "G", "H",
		"I", "J", "K", "L",
		"M", "N", "O", "N",
	}

	var tiles []Tile
	for i, word := range allWords {
		tiles = append(tiles, Tile{
			ID:       i + 1,
			Word:     word,
			Selected: false,
			GroupID:  0,
		})
	}

	// Shuffle tiles
	rand.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})

	return &GameState{
		Tiles:          tiles,
		MistakesLeft:   4,
		Date:           time.Now().Format("January 2, 2006"),
		Message:        "Create four groups of four!",
		SelectedTiles:  []int{},
		CorrectGroups:  correctGroups,
		SolvedGroups:   [][]string{},
		CurrentGroupID: 0,
	}
}

func containsSameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Convert both slices to lowercase for case-insensitive comparison
	aLower := make([]string, len(a))
	bLower := make([]string, len(b))

	for i, val := range a {
		aLower[i] = strings.ToLower(val)
	}

	for i, val := range b {
		bLower[i] = strings.ToLower(val)
	}

	// Check if each element in a is in b
	for _, val := range aLower {
		found := false
		for _, val2 := range bLower {
			if val == val2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
