const errorLabelClassName = "error-message";
const validations = [
    {
        id: 'firstname',
        message: 'Imię powinno zawierać przynajmniej 2 znaki, wszystkie powinny pochodzić z polskiego alfabetu i tylko pierwszy znak powinien być dużą literą.',
    },
    {
        id: 'lastname',
        message: 'Nazwisko powinno zawierać przynajmniej 2 znaki, wszystkie powinny pochodzić z polskiego alfabetu i tylko pierwszy znak powinien być dużą literą.',
    },
    {
        id: 'login',
        message: 'Nazwa użytkownika powinna składać się wyłącznie z małych liter i mieć długość od 3 do 12 znaków.',
        callback: async (input) => {
            const inputValue = input.value;
            const res = await fetch(`https://infinite-hamlet-29399.herokuapp.com/check/${inputValue}`);
            if (input.value !== inputValue) {
                throw new Error('Value has changed');
            }
            if (res.status !== 200) {
                return 'Nie można obecnie sprawdzić czy podana nazwa użytkownika jest dostępna.';
            }
            const value = await res.json()
            return value[inputValue] !== 'available'
                ? 'Podana nazwa użytkownika jest niedostępna.'
                : null;
        }
    },
    {
        id: 'password',
        message: 'Hasło powinno składać się z przynajmniej 8 znaków.'
    },
    {
        id: 'passwordConfirmation',
        message: 'Hasło powinno składać się z przynajmniej 8 znaków.',
        callback: (input) => {
            const password = document.getElementById('password');
            return password.value !== input.value
                ? 'Hasła powinny się pokrywać.'
                : null;
        }
    }
]

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

async function getValidationMessage(validation, input) {
    if (input.validity.valid) {
        if (validation.callback) {
            return await validation.callback(input);
        }
    } else {
        return validation.message;
    }
}

async function validateInput(validation, input) {
    removeErrorLabel(input);
    let message;
    try {
        message = await getValidationMessage(validation, input)
    } catch(e) {
        return;
    }
    if (message) {
        addErrorLabel(input, message);
    }
}

function attachEvents() {    
    validations.forEach((validation) => {
        const input = document.getElementById(validation.id);
        input.addEventListener('input', async (e) => validateInput(validation, e.target));
    })

    const form = document.getElementById('signUpForm');
    form.noValidate = true;
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const validateInputs = [
            ...validations,
            {
                id: 'sex',
                message: 'Pole jest obowiązkowe.'
            }
        ].map(async (validation) => {
            const input = document.getElementById(validation.id);
            await validateInput(validation, input);
        })
        
        await Promise.all(validateInputs)

        const errorLabels = form.getElementsByClassName(errorLabelClassName);
        if (errorLabels.length === 0) {
            form.submit();
        }
    });
}

attachEvents();
