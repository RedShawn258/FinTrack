// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

// Custom command for login
Cypress.Commands.add('login', (identifier = 'durgamaheshboppani@gmail.com', password = 'Mahesh@1078') => {
  // Visit login page
  cy.visit('/login');
  
  // Enter login credentials - updated selectors based on actual DOM
  cy.get('input[placeholder="Username/Email"]').type(identifier);
  cy.get('input[placeholder="Password"]').type(password);
  
  // Submit login form
  cy.get('button[type="submit"]').click();
  
  // Verify redirection to dashboard
  cy.url().should('include', '/dashboard');
});

// Custom command to set authenticated state without UI
Cypress.Commands.add('loginByAuth', (token = 'test-token', username = 'testuser') => {
  cy.window().then((win) => {
    win.localStorage.setItem('user', JSON.stringify({
      token,
      username
    }));
  });
});