
function browsePrint() {
    // Force graphs load then trigger print
    graphHandleQueue(true).then(function () {
        window.print();
    });
}
