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
    const login = document.getElementById("login");
    login.addEventListener("input", (e) => {
        const loginValue = e.target.value;
        removeErrorLabel(login);
        if (login.validity.valid) {
            fetch(`https://infinite-hamlet-29399.herokuapp.com/check/${loginValue}`).then((res) => {
                if (login.value === loginValue) {
                    if (res.status === 200) {
                        res.json().then((value) => {
                            if (value[loginValue] !== 'available') {
                                addErrorLabel(login, 'Podana nazwa użytkownika jest niedostępna.');
                            }
                        });
                    } else {
                        addErrorLabel(login, 'Nie można obecnie sprawdzić czy podana nazwa użytkownika jest dostępna.');
                    }
                }
            });
        } else {
            addErrorLabel(login, 'Nazwa użytkownika powinna składać się wyłącznie z małych liter i mięć długość od 3 do 12 znaków.');
        }
    });
}

attachEvents();
