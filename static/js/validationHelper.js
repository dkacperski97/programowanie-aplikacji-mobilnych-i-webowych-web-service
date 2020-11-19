const errorLabelClassName = 'error-message';

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

export async function validateControl(validation, control, validateWithoutCallbacks) {
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

export default function attachEvents(inputsValidations, formName, validateWithoutCallbacks) {
	inputsValidations.forEach((validation) => {
		const input = document.getElementById(validation.id);
		if (input !== null) {
			input.addEventListener('input', async (e) =>
				validateControl(validation, e.target, validateWithoutCallbacks)
			);
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
