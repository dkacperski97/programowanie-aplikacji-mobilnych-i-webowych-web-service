const errorLabelClassName = 'error-message';

const inputsValidations = [
	{
		id: 'firstname',
		message:
			'Imię powinno zawierać przynajmniej 2 znaki, wszystkie powinny pochodzić z polskiego alfabetu i tylko pierwszy znak powinien być dużą literą.',
	},
	{
		id: 'lastname',
		message:
			'Nazwisko powinno zawierać przynajmniej 2 znaki, wszystkie powinny pochodzić z polskiego alfabetu i tylko pierwszy znak powinien być dużą literą.',
	},
	{
		id: 'login',
		message:
			'Nazwa użytkownika powinna składać się wyłącznie z małych liter i mieć długość od 3 do 12 znaków.',
		callback: async (input) => {
			const inputValue = input.value;
			const res = await fetch(`/check/${inputValue}`);
			if (input.value !== inputValue) {
				throw new Error('Value has changed');
			}
			if (res.status !== 200) {
				return 'Nie można obecnie sprawdzić czy podana nazwa użytkownika jest dostępna.';
			}
			const value = await res.json();
			return value[inputValue] !== 'available' ? 'Podana nazwa użytkownika jest niedostępna.' : null;
		},
	},
	{
		id: 'email',
		message: 'Podano niepoprawny adres e-mail.',
	},
	{
		id: 'password',
		message: 'Hasło powinno składać się z przynajmniej 8 znaków.',
	},
	{
		id: 'passwordConfirmation',
		message: 'Hasło powinno składać się z przynajmniej 8 znaków.',
		callback: (input) => {
			const password = document.getElementById('password');
			return password.value !== input.value ? 'Hasła powinny się pokrywać.' : null;
		},
	},
	{
		id: 'address',
		message: 'Pole adres jest wymagane.',
	},
];

function addErrorLabel(control, error) {
	const label = document.createElement('label');
	label.setAttribute('for', control.id);
	label.classList.add(errorLabelClassName);
	label.textContent = error;
	control.after(label);
}

function removeErrorLabel(control) {
	const label = control.nextElementSibling;
	if (label !== null && label.classList.contains(errorLabelClassName)) {
		label.remove();
	}
}

async function getValidationMessage(validation, control, validateWithoutCallbacks) {
	if (control.validity.valid) {
		if (validation.callback && !validateWithoutCallbacks) {
			return await validation.callback(control);
		}
	} else {
		return validation.message;
	}
}

async function validateControl(validation, control, validateWithoutCallbacks) {
	removeErrorLabel(control);
	let message;
	try {
		message = await getValidationMessage(validation, control, validateWithoutCallbacks);
	} catch (e) {
		return;
	}
	if (message) {
		addErrorLabel(control, message);
	}
}

export default function attachEvents(formName, validateWithoutCallbacks) {
	inputsValidations.forEach((validation) => {
        const input = document.getElementById(validation.id);
        if (input !== null) {
            input.addEventListener('input', async (e) => validateControl(validation, e.target, validateWithoutCallbacks));
        }
	});

	const form = document.getElementById(formName);
	form.noValidate = true;
	form.addEventListener('submit', async (e) => {
		e.preventDefault();
		const validations = inputsValidations.map(async (validation) => {
            const control = document.getElementById(validation.id);
            if (control) {
                await validateControl(validation, control, validateWithoutCallbacks);
            }
		});

		await Promise.all(validations);

		const errorLabels = form.getElementsByClassName(errorLabelClassName);
		if (errorLabels.length === 0) {
			form.submit();
		}
	});
}