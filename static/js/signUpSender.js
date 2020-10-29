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
			const res = await fetch(`https://infinite-hamlet-29399.herokuapp.com/check/${inputValue}`);
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

async function getValidationMessage(validation, control) {
	if (control.validity.valid) {
		if (validation.callback) {
			return await validation.callback(control);
		}
	} else {
		return validation.message;
	}
}

async function validateControl(validation, control) {
	removeErrorLabel(control);
	let message;
	try {
		message = await getValidationMessage(validation, control);
	} catch (e) {
		return;
	}
	if (message) {
		addErrorLabel(control, message);
	}
}

function attachEvents() {
	inputsValidations.forEach((validation) => {
		const input = document.getElementById(validation.id);
		input.addEventListener('input', async (e) => validateControl(validation, e.target));
	});

	const form = document.getElementById('signUpForm');
	form.noValidate = true;
	form.addEventListener('submit', async (e) => {
		e.preventDefault();
		const validations = [
			...inputsValidations,
			{
				id: 'sex',
				message: 'Pole jest obowiązkowe.',
			},
		].map(async (validation) => {
			const control = document.getElementById(validation.id);
			await validateControl(validation, control);
		});

		await Promise.all(validations);

		const errorLabels = form.getElementsByClassName(errorLabelClassName);
		if (errorLabels.length === 0) {
			form.submit();
		}
	});
}

attachEvents();
