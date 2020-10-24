const errorLabelClassName = "error-message";

function addErrorLabel(input, error) {
    const label = document.createElement("label");
    label.setAttribute('for', input.id)
    label.classList.add(errorLabelClassName);
    label.textContent = error;
    input.after(label);
}

function removeErrorLabel(input) {
    const label = input.nextElementSibling;
    if (label !== null && label.classList.contains(errorLabelClassName)) {
        label.remove();
    }
}

function attachEvents() {
    const username = document.getElementById("username");
    username.addEventListener("input", (e) => {
        const usernameValue = e.target.value;
        removeErrorLabel(username);
        if (username.value) {
            fetch(`https://infinite-hamlet-29399.herokuapp.com/check/${usernameValue}`).then((res) => {
                if (username.value === usernameValue) {
                    if (res.status === 200) {
                        res.json().then((value) => {
                            if (value[usernameValue] !== 'available') {
                                addErrorLabel(username, 'Podana nazwa użytkownika jest niedostępna.');
                            }
                        });
                    } else {
                        addErrorLabel(username, 'Nie można obecnie sprawdzić czy podana nazwa użytkownika jest dostępna.');
                    }
                }
            });
        } else {
            addErrorLabel(username, 'Nazwa użytkownika jest wymagana.');
        }
    });
}

attachEvents();
