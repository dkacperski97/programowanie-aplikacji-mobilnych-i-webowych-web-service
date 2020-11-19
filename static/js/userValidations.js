import { validateControl } from "./validationHelper.js";

export default [
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
		callback: () => {
			const passwordConfirmationId = 'passwordConfirmation';
			const passwordConfirmation = document.getElementById(passwordConfirmationId);
			const validation = userValidations.find(validation => validation.id === passwordConfirmationId);
			if (passwordConfirmation?.value.length > 0 && validation) {
				validateControl(validation, passwordConfirmation, false);
				return null
			}
		}
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
