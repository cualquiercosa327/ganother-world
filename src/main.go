package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"os"
	"sort"
)

//GameState used to save and load a game state
type GameState struct {
	vm    VMState
	video Video
}

// video is a global variable that needs to implement the Renderer interface
var video Video
var gameState GameState

func main() {
	Info("# GOTHER WORLD vDEV")

	noVideoOutput := flag.Bool("t", false, "Use Text only output (no SDL needed)")
	debug := flag.Bool("d", false, "Enable Debug Mode")
	startPart := flag.Int("p", 1, "Game part to start from (0-9)")
	flag.Parse()

	Info("# KEYBOARD MAPPING:")
	Info("- L: Load State")
	Info("- S: Save State")

	if *debug == false {
		SetLogLevel(LEVEL_INFO)
	}
	video = initVideo(*noVideoOutput)

	Info("- load memlist.bin")
	data := readFile("./assets/memlist.bin")
	resourceMap, resourceStatistics := unmarshallingMemlistBin(data)
	printResourceStats(resourceStatistics)

	bankFilesMap := createBankMap("./assets/")
	gameParts := getGameParts()
	assets := Assets{
		memList:         resourceMap,
		gameParts:       gameParts,
		bank:            bankFilesMap,
		loadedResources: make(map[int][]uint8),
	}

	Info("- create state")
	vmState := createNewState(assets)

	Info("- setup game")
	loadGamePart(&vmState, GAME_PART_ID_1+*startPart)

	//start main loop
	keyPresses := uint32(0)
	for i := 0; keyPresses&KEY_ESC == 0; i++ {
		/*if i%30 == rand.Intn(30) {
			loadGamePart(&vmState, GAME_PART_ID_1+rand.Intn(9))
		}*/

		keyPresses = video.eventLoop(i)
		vmState.mainLoop(keyPresses)

		if keyPresses&KEY_SAVE > 0 {
			Info("SAVE STATE")
			gameState = GameState{vmState, video}
		}
		if gameState.vm.gamePart > 0 && keyPresses&KEY_LOAD > 0 {
			Info("LOAD STATE")
			vmState.loadGameParts(gameState.vm.gamePart)
			vmState.variables = gameState.vm.variables
			vmState.channelPC = gameState.vm.channelPC
			vmState.nextLoopChannelPC = gameState.vm.nextLoopChannelPC
			vmState.channelPaused = gameState.vm.channelPaused
			vmState.stackCalls = gameState.vm.stackCalls
			video = gameState.video
		}

		if vmState.loadNextPart > 0 {
			Info("- load next part %d", vmState.loadNextPart)
			loadGamePart(&vmState, vmState.loadNextPart)
		}

		Debug("exit=%d", keyPresses)
		//game run at approx 25 fps
		time.Sleep(20 * time.Millisecond)
	}

	video.shutdown()
}

func loadGamePart(vmState *VMState, partID int) {
	vmState.setupGamePart(partID)
	videoAssets := vmState.buildVideoAssets()
	video.updateGamePart(videoAssets)
}

func readFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		Error("File reading error %v", err)
		os.Exit(1)
	}
	return data
}

func createBankMap(assetPath string) map[int][]byte {
	bankFilesMap := make(map[int][]byte)
	for i := 0x01; i < 0x0e; i++ {
		name := fmt.Sprintf("%sbank%02x", assetPath, i)
		Debug("- load file %s", name)
		entry := readFile(name)
		bankFilesMap[i] = entry
	}
	return bankFilesMap
}

func printResourceStats(memlistStatistic MemlistStatistic) {
	Debug("Total # resources: %d", memlistStatistic.entryCount)
	Debug("Compressed       : %d", memlistStatistic.compressedEntries)
	Debug("Uncompressed     : %d", memlistStatistic.entryCount-memlistStatistic.compressedEntries)
	var compressionRatio float64 = 100 / float64(memlistStatistic.entryCount) * float64(memlistStatistic.compressedEntries)
	Debug("Note: %.0f%% of resources are compressed.", math.Round(compressionRatio))
	Debug("Total size (uncompressed) : %d bytes.", memlistStatistic.sizeUncompressed)
	Debug("Total size (compressed)   : %d bytes.", memlistStatistic.sizeCompressed)
	var compressionGain float64 = 100 * (1 - float64(memlistStatistic.sizeCompressed)/float64(memlistStatistic.sizeUncompressed))
	Debug("Note: Overall compression gain is : %.0f%%.", math.Round(compressionGain))

	sortedKeys := sortedKeys(memlistStatistic.resourceTypeMap)
	for i := 0; i < len(sortedKeys); i++ {
		k := sortedKeys[i]
		resourceName := getResourceTypeName(k)
		if len(resourceName) < 1 {
			resourceName = fmt.Sprintf("RT_UNKOWNN_%d", k)
		}
		Debug("Total %20s, files: %d", resourceName, memlistStatistic.resourceTypeMap[k])
	}
}

func sortedKeys(m map[int]int) []int {
	keys := make([]int, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	return keys
}
