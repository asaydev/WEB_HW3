// const add_form = document.getElementById("add_form");
// const first_num = document.getElementById("fi").value;
// const second_num = document.getElementById("si").value;
// const find_form = document.getElementById("find_form");
const register_form = document.getElementById("reg-form");
const login_form = document.getElementById("login-form");

function registerr() {
    alert('bye');
    document.getElementById("reg-form").addEventListener("submit", (ev => {
        alert('hi');
        // e.preventDefault(); // ?
        const request = new XMLHttpRequest();
        request.open("post", "http://192.168.1.101/api/signup", true);
        console.log(`email=${document.getElementById("inputEmail4").value}`);
        request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
        request.onreadystatechange = function () { // Call a function when the state changes.
            if (this.readyState === XMLHttpRequest.DONE && this.status === 201) {
                location.replace("http://192.168.1.101");
                alert(request.responseText)
            }
            if (this.readyState === XMLHttpRequest.DONE && this.status !== 201) {
                alert(this.status)
            }
        };
        request.send(`email=${document.getElementById("inputEmail4").value}&password=${document.getElementById("inputPassword4").value}`);
    }))
}

function loginn() {
    document.getElementById("login-form").addEventListener("submit", (ev => {
        e.preventDefault(); // ?
        let request = new XMLHttpRequest();
        request.open("post", "http://192.168.1.101/api/signup", true);
        request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
        request.onreadystatechange = function () { // Call a function when the state changes.
            if (this.readyState === XMLHttpRequest.DONE && this.status === 200) {

            }
            if (this.readyState === XMLHttpRequest.DONE && this.status !== 200) {
                // not 200
            }
            alert(this.status)
        };
        request.send(`email=${document.getElementById("inputEmail3").value}&password=${document.getElementById("inputPassword3").value}`);
    }))
}

