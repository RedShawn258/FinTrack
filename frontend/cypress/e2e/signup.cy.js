describe('Signup Page Tests', () => {
  beforeEach(() => {
    // Visit the signup page before each test
    cy.visit('/signup')
  })

  it('should display signup form with required elements', () => {
    // Check if signup form elements are present - updated selectors
    cy.get('input[placeholder="Username"]').should('exist')
    cy.get('input[placeholder="Email"]').should('exist')
    cy.get('input[placeholder="Password"]').should('exist')
    cy.get('input[placeholder="Confirm Password"]').should('exist')
    cy.get('button[type="submit"]').should('exist')
  })

  it('should show validation errors for empty fields', () => {
    // Click signup without entering any data
    cy.get('button[type="submit"]').click()
    
    // Check for validation errors
    cy.get('input:invalid').should('exist')
  })

  it('should show error for password mismatch', () => {
    // Fill out form with mismatched passwords
    cy.get('input[placeholder="Username"]').type('testuser')
    cy.get('input[placeholder="Email"]').type('test@example.com')
    cy.get('input[placeholder="Password"]').type('Password123!')
    cy.get('input[placeholder="Confirm Password"]').type('DifferentPassword123!')
    cy.get('button[type="submit"]').click()
    
    // Since the app uses alert for error messages, intercept it
    cy.on('window:alert', (text) => {
      expect(text).to.contains('Passwords do not match')
    })
  })

  it('should navigate to login page when clicking login link', () => {
    // Find and click the login link
    cy.contains('Already have an account?').parent().find('a').click()
    
    // Verify redirect to login page
    cy.url().should('include', '/login')
  })
}); 