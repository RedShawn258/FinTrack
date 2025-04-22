describe('Insights Page', () => {
  beforeEach(() => {
    // Mock the login state using custom command
    cy.loginByAuth('test-token', 'testuser');

    // Visit the insights page
    cy.visit('/insights');
  });

  it('displays insights page with analytics tab by default', () => {
    cy.get('.insights-header').should('contain', 'Insights & Achievements');
    cy.get('.spending-trends-card').should('be.visible');
    cy.get('.category-breakdown-card').should('be.visible');
    cy.get('.savings-card').should('be.visible');
  });

  it('switches between analytics and achievements tabs', () => {
    // Check analytics tab
    cy.get('.tab-button').contains('Analytics').should('have.class', 'active');
    cy.get('.spending-trends-card').should('be.visible');

    // Switch to achievements tab
    cy.get('.tab-button').contains('Achievements').click();
    cy.get('.tab-button').contains('Achievements').should('have.class', 'active');
    cy.get('.achievements-grid').should('be.visible');
    cy.get('.achievement-card').should('have.length.at.least', 1);
  });

  it('displays charts when data is available', () => {
    cy.get('.chart-container').should('have.length.at.least', 3);
    cy.get('.spending-trends-card .chart-container').should('be.visible');
    cy.get('.category-breakdown-card .chart-container').should('be.visible');
    cy.get('.savings-card .chart-container').should('be.visible');
  });

  it('displays no data message when data is not available', () => {
    // Clear the session storage to simulate no data
    cy.window().then((win) => {
      win.sessionStorage.removeItem('dashboardData');
    });
    cy.reload();

    cy.get('.no-data-message').should('be.visible');
  });

  it('navigates back to dashboard', () => {
    cy.get('.back-button').click();
    cy.url().should('include', '/dashboard');
  });

  it('handles logout', () => {
    cy.get('.logout-icon-button').click();
    cy.url().should('equal', Cypress.config().baseUrl + '/');
    cy.window().then((win) => {
      expect(win.localStorage.getItem('user')).to.be.null;
    });
  });
}); 