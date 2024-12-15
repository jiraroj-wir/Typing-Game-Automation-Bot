package main

import (
	"fmt"
    "image"
    "image/color"
	"image/png"
    //"strings"
	"os"

	"github.com/kbinani/screenshot"
    "github.com/disintegration/imaging"
    "github.com/tiagomelo/go-ocr/ocr"
    //"github.com/go-vgo/robotgo"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func cropScreenshot(inputPath, outputPath string, cropRatio float64) error {
    file, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    img, err := png.Decode(file)
    if err != nil {
        return err
    }

    // Calculate the cropping dimensions
    bounds := img.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()

    // Calculate the new crop rectangle
    cropHeight := int(cropRatio * float64(height))
    cropRect := image.Rect(0, cropHeight, width, height)

    // Crop the image
    croppedImage := img.(SubImager).SubImage(cropRect)

    // Save the cropped image
    croppedImageFile, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer croppedImageFile.Close()

    if err := png.Encode(croppedImageFile, croppedImage); err != nil {
        return err
    }

    return nil
}

func convertToBinary(inputPath, outputPath string, threshold, contrastFactor float64) error {
	// Open the input image
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Convert to grayscale
	grayImg := imaging.Grayscale(img)

	// Resize the image to improve resolution (upscale by 2x)
	upscaledImg := imaging.Resize(grayImg, grayImg.Bounds().Dx()*2, grayImg.Bounds().Dy()*2, imaging.Lanczos)

	// Apply binary threshold
	binaryImg := imaging.AdjustFunc(upscaledImg, func(c color.NRGBA) color.NRGBA {
		brightness := float64(c.R) / 255.0
		if brightness > threshold {
			return color.NRGBA{R: 255, G: 255, B: 255, A: c.A}
		}
		return color.NRGBA{R: 0, G: 0, B: 0, A: c.A}
	})

	// Increase contrast
	finalImg := imaging.AdjustContrast(binaryImg, contrastFactor)

	// Save the enhanced binary image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, finalImg)
	if err != nil {
		return fmt.Errorf("failed to encode enhanced binary image: %v", err)
	}

	return nil
}

func main() {
	err := os.MkdirAll("screenshots", os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return
	}

	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		fmt.Printf("Failed to capture screen: %v\n", err)
		return
	}

	file, err := os.Create("screenshots/screenshot.png")
	if err != nil {
		fmt.Printf("Failed to save screenshot: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("Failed to encode screenshot: %v\n", err)
		return
	}


    err = cropScreenshot("screenshots/screenshot.png", "screenshots/cropped.png", 90.0/900.0)
    if err != nil {
        fmt.Println("Error cropping image:", err)
        return
    }

    err = convertToBinary("screenshots/cropped.png", "screenshots/binary.png", 0.5, 80.0)
    if err != nil {
        fmt.Println("Error enhancing binary image:", err)
        return
    }

    const TesseractPath = "/opt/homebrew/bin/tesseract"
	t, err := ocr.New(ocr.TesseractPath(TesseractPath))
	if err != nil {
        fmt.Println("Error finding tesseract:", err)
        return
	}
	extractedText, err := t.TextFromImageFile("screenshots/binary.png")
	if err != nil {
        fmt.Println("Error extracting text from binary: ", err)
        return
	}

    fmt.Println(extractedText)
    /*
    // Split the extracted text into slices of words based on spaces or newlines
    words := strings.Fields(extractedText)

    
    // Type each word followed by pressing "Enter"
    for _, word := range words {
        robotgo.TypeStr(word)
        robotgo.KeyTap("enter")
    }

    // typeText := strings.Join(words, " ")
    // robotgo.TypeStr(typeText)

    */
}

