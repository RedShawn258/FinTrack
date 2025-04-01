describe('Login Page Tests', () => {
  // Define test data
  const validUser = {
    email: 'durgamaheshboppani@gmail.com',
    password: 'Mahesh@1078'
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
    // Check if login form elements are present
    cy.get('input[type="text"]').should('exist')
    cy.get('input[type="password"]').should('exist')
    cy.get('button').should('exist')
  })

  it('should show validation errors for empty fields', () => {
    // Click login without entering any data
    cy.get('button').click()
    
    // Check for validation error messages
    cy.get('input[type="text"]:invalid').should('exist')
    cy.get('input[type="password"]:invalid').should('exist')
  })

  it('should show error for invalid email format', () => {
    // Enter invalid email format
    cy.get('input[type="text"]').type('invalidemail')
    cy.get('input[type="password"]').type('password123')
    cy.get('button').click()
    
    // Check for alert message
    cy.on('window:alert', (text) => {
      expect(text).to.contains('Invalid credentials')
    })
  })

  it('should show error for invalid credentials', () => {
    // Enter invalid credentials
    cy.get('input[type="text"]').type(invalidUser.email)
    cy.get('input[type="password"]').type(invalidUser.password)
    cy.get('button').click()
    
    // Check for alert message
    cy.on('window:alert', (text) => {
      expect(text).to.contains('Invalid credentials')
    })
  })

  it('should successfully login with valid credentials and show dashboard data', () => {
    // Intercept the transactions API call
    cy.intercept('GET', 'http://localhost:8080/api/v1/transactions').as('getTransactions')
    
    // Enter valid credentials
    cy.get('input[type="text"]').type(validUser.email)
    cy.get('input[type="password"]').type(validUser.password)
    cy.get('button').click()
    
    // Wait for the dashboard URL
    cy.url().should('include', '/dashboard')

    // Wait for the API response
    cy.wait('@getTransactions')
    
    // Verify dashboard elements are visible with increased timeout
    cy.get('.dashboard-container', { timeout: 10000 }).should('exist')
    
    // Verify Total Expenses header is visible
    cy.contains('h2', 'Total Expenses', { timeout: 10000 }).should('be.visible')
  })

  it('should successfully logout and redirect to login page', () => {
    // First login
    cy.get('input[type="text"]').type(validUser.email)
    cy.get('input[type="password"]').type(validUser.password)
    cy.get('button').click()
    
    // Wait for dashboard to load
    cy.url().should('include', '/dashboard')
    
    // Click logout icon button
    cy.get('.logout-icon-button').click()
    
    // Verify redirect to login page
    cy.url().should('include', '/login')
    
    // Verify login form is visible
    cy.get('input[type="text"]').should('exist')
    cy.get('input[type="password"]').should('exist')
    cy.get('button').should('exist')
  })
})