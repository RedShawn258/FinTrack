describe('Budget Planning', () => {
  beforeEach(() => {
    // Mock the login state using custom command
    cy.loginByAuth('test-token', 'testuser');

    // Visit the dashboard page which contains budget functionality
    cy.visit('/dashboard');
  });

  it('displays budget overview section', () => {
    // Check if budget overview section is visible
    cy.contains('Budget Overview').should('be.visible');
    cy.contains('Set New Budget').should('be.visible');
  });

  it('has form fields for setting budgets', () => {
    cy.get('.budget-field').should('exist');
    cy.get('input[placeholder="Select or type category"]').should('exist');
    cy.get('input[placeholder="Enter amount"]').should('exist');
    cy.get('.set-budget-button').should('be.visible');
  });

  it('displays current budgets section', () => {
    cy.contains('Current Budgets').should('be.visible');
  });
}); 