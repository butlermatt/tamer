function getActive() {
    fetch('Active').then(function(response) {
        return response.json()
    }).then(function(data) {
       for (var i = 0; i < data.length; i++) {
           console.log(data[i]);
       }
    }).catch(function(error) {
        console.log('Error fetching active planes: ' + error)
    });
}

function main() {
    var cb = window.setInterval(getActive, 5 * 1000)
}