// script.js makes the "Download .txt" button work.
// It takes the ASCII art shown in <pre id="result"> and saves it as a text file.
// Everything happens in the browser: the server is not called again.

const downloadButton = document.getElementById("download");
const result = document.getElementById("result");

// The button only exists after a result was generated
if (downloadButton && result) {
    downloadButton.addEventListener("click", function () {
        // 1. Put the art in a Blob: a file that lives in the browser's memory
        const file = new Blob([result.textContent], { type: "text/plain" });

        // 2. Give the Blob a temporary URL, and point a hidden link at it
        const url = URL.createObjectURL(file);
        const link = document.createElement("a");
        link.href = url;
        link.download = "ascii-art.txt"; // the name of the saved file

        // 3. Click the link to start the download, then free the URL
        link.click();
        URL.revokeObjectURL(url);
    });
}
