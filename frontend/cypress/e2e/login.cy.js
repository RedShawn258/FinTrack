describe('Login Page Tests', () => {
  // Define test data
  const validUser = {
    email: Cypress.env('TEST_USER_EMAIL') || 'durgamaheshboppani@gmail.com',
    password: Cypress.env('TEST_USER_PASSWORD') || 'Mahesh@1078'
  }

  const invalidUser = {
    email: 'wrong@example.com',
    password: 'wrongpassword'
  }

  beforeEach(() => {
    // Visit the login page before each test
    cy.visit('/login')
  })

  it('should display login form with required elements', () => {
    // Check if login form elements are present (updated selectors)
    cy.get('input[placeholder="Username/Email"]').should('exist')
    cy.get('input[placeholder="Password"]').should('exist')
    cy.get('button[type="submit"]').should('exist')
  })

  it('should show validation errors for empty fields', () => {
    // Click login without entering any data
    cy.get('button[type="submit"]').click()
    
    // Check for validation error messages
    cy.get('input:invalid').should('exist')
  })

  it('should show error for invalid email format', () => {
    // Enter invalid email format
    cy.get('input[placeholder="Username/Email"]').type('invalidemail')
    cy.get('input[placeholder="Password"]').type('password123')
    cy.get('button[type="submit"]').click()
    
    // Check for error message
    cy.get('.error-message').should('be.visible')
  })

  it('should show error for invalid credentials', () => {
    // Enter invalid credentials
    cy.get('input[placeholder="Username/Email"]').type(invalidUser.email)
    cy.get('input[placeholder="Password"]').type(invalidUser.password)
    cy.get('button[type="submit"]').click()
    
    // Check for error message
    cy.get('.error-message').should('be.visible')
  })

  it('should successfully login with valid credentials and show dashboard data', () => {
    // Use custom login command
    cy.login(validUser.email, validUser.password)
    
    // Verify dashboard elements are visible
    cy.contains('Total Expenses').should('be.visible', { timeout: 10000 })
  })

  it('should successfully logout and redirect to login page', () => {
    // First login
    cy.login(validUser.email, validUser.password)
    
    // Click logout icon button
    cy.get('.logout-icon-button').click()
    
    // Verify redirect to home/landing page (not login)
    cy.url().should('equal', Cypress.config().baseUrl + '/')
    
    // Verify we've logged out by checking for landing page elements
    cy.contains('FinTrack').should('be.visible')
  })
})