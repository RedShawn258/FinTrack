describe('Profile Page', () => {
  beforeEach(() => {
    // Mock the login state using custom command
    cy.loginByAuth('test-token', 'testuser');

    // Visit the profile page directly
    cy.visit('/profile');
  });

  it('displays profile page with user information', () => {
    cy.get('.profile-header').should('contain', 'User Profile');
    cy.get('form').should('exist');
  });

  it('allows navigating back to dashboard', () => {
    cy.get('.back-button').contains('Dashboard').should('be.visible');
    cy.get('.back-button').click();
    cy.url().should('include', '/dashboard');
  });

  it('handles logout correctly', () => {
    cy.get('.logout-icon-button').click();
    cy.url().should('equal', Cypress.config().baseUrl + '/');
  });
}); 