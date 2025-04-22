describe('Dashboard Page', () => {
  beforeEach(() => {
    // Mock the login state using custom command
    cy.loginByAuth('test-token', 'testuser');

    // Visit the dashboard page
    cy.visit('/dashboard');
  });

  it('displays dashboard with key financial data', () => {
    // Check that dashboard elements are present
    cy.get('.dashboard-container').should('exist');
    cy.contains('Total Expenses').should('be.visible');
  });

  it('displays add expense form', () => {
    // Check if the add expense form is visible
    cy.contains('Add New Expense').should('be.visible');
    cy.get('input[type="text"]').should('exist');
    cy.get('input[type="number"]').should('exist');
    cy.get('input[type="date"]').should('exist');
    cy.get('button').contains('Add Expense').should('exist');
  });

  it('displays budget overview section', () => {
    // Check if budget overview section is visible
    cy.contains('Budget Overview').should('be.visible');
    cy.contains('Set New Budget').should('be.visible');
  });

  it('navigates to insights page', () => {
    cy.get('.insights-icon-button').click();
    cy.url().should('include', '/insights');
  });

  it('navigates to profile page', () => {
    cy.get('.profile-icon-button').click();
    cy.url().should('include', '/profile');
  });

  it('handles logout', () => {
    cy.get('.logout-icon-button').click();
    cy.url().should('equal', Cypress.config().baseUrl + '/');
  });
}); 