# **FinTrack - Personal Finance & Budgeting Platform**

## **Overview**
FinTrack is a personal finance and budgeting platform that helps users gain real-time insights into their spending habits, set budgets, track expenses, and earn rewards for financial progress. By integrating gamification elements like badges, points, and leaderboards, FinTrack makes financial management interactive and engaging.

---

## **Features**
- **User Authentication:** Secure Signup, Login, and Forgot Password functionality.
- **Expense Tracking:** Add, edit, and delete expenses under different categories.
- **Budget Management:** Set and adjust monthly budgets with real-time monitoring.
- **Analytics & Insights:** Visual representation of spending patterns.
- **Gamification Elements:** Earn badges and rewards for savings milestones.
- **Alerts & Notifications:** Get reminders for bill payments and budget thresholds.
- **Data Visualization:** Interactive charts and graphs for expense analysis.
- **Responsive Design:** Optimized for all device sizes.
- **Error Handling:** Robust validation and error management.
- **Landing Pages:** Informative landing pages with testimonials and feature highlights.
- **AWS Infrastructure:** Cloud-based deployment with Infrastructure as Code.
- **Database Optimization:** Enhanced performance and secure access patterns.
- **Expense Forecasting:** AI-powered expense predictions and financial planning.
- **Advanced Unit Testing:** Comprehensive test coverage for all backend features.

---

## **Project Summary**

FinTrack has been successfully completed with all planned features implemented and tested. The project delivers a comprehensive personal finance solution with emphasis on user experience, performance, and security.

### **Key Accomplishments**

✅ Implemented complete expense tracking and budget management system
✅ Created a gamification system to increase user engagement
✅ Developed advanced analytics and forecasting capabilities
✅ Built responsive and intuitive UI/UX across all device types
✅ Established comprehensive testing framework with high test coverage
✅ Optimized database performance and security
✅ Deployed on AWS infrastructure with monitoring and scaling capabilities

---

## **Technology Stack**

### **Backend**
- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Database**: MySQL (AWS RDS)
- **ORM**: GORM
- **Authentication**: JWT
- **Deployment**: AWS (EC2, RDS, S3)
- **Testing**: Go Testing Framework

### **Frontend**
- **Framework**: React.js
- **State Management**: Redux
- **UI Library**: Material-UI
- **Charts**: Chart.js, D3.js
- **Testing**: Jest, React Testing Library, Cypress
- **CSS**: Styled Components, SASS

### **Infrastructure**
- **Cloud Provider**: AWS
- **CI/CD**: GitHub Actions
- **Containerization**: Docker
- **Infrastructure as Code**: Terraform
- **Monitoring**: CloudWatch, Prometheus
- **Security**: AWS WAF, AWS Secrets Manager

---

## **Project Architecture**

### **Backend Architecture**
The backend follows a clean architecture with the following components:
- **Handlers**: HTTP request handlers for the REST API
- **Services**: Business logic layer that implements app features
- **Repositories**: Data access layer that interacts with the database
- **Models**: Data models that represent database entities
- **Middleware**: Authentication, logging, and error handling
- **Routes**: API route definitions and organization

### **Frontend Architecture**
The frontend follows a component-based architecture:
- **Components**: Reusable UI elements
- **Pages**: Complete page layouts
- **Services**: API interaction logic
- **Store**: State management with Redux
- **Hooks**: Custom React hooks for business logic
- **Tests**: Unit, integration, and E2E tests

---

## **Setup & Installation**

### **Backend Setup**
1. Clone the repository:
   ```bash
   git clone https://github.com/RedShawn258/FinTrack.git
   cd FinTrack/backend
   ```
2. Install dependencies:
   ```bash
   go mod init github.com/RedShawn258/FinTrack/backend
   go mod tidy
   ```
3. Configure environment:
   ```bash
   cp .env.example .env
   # Update the .env file with your configurations
   ```
4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

### **Frontend Setup**
1. Navigate to the frontend directory:
   ```bash
   cd FinTrack/frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Configure environment:
   ```bash
   cp .env.example .env
   # Update the .env file with your configurations
   ```
4. Start the frontend application:
   ```bash
   npm start
   ```

### **Running Tests**
1. Frontend Tests:
   ```bash
   # Unit tests
   npm test
   
   # Coverage report
   npm test -- --coverage
   ```
2. Cypress E2E Tests:
   ```bash
   # Open Cypress test runner
   npm run cypress:open
   
   # Run tests headlessly
   npm run cypress:run
   ```
3. Backend Tests:
   ```bash
   # Run all tests
   cd backend
   ./run_all_tests.sh
   
   # Run specific tests
   go test ./internal/handlers_test
   ```

---

## **API Documentation**

### **Authentication Endpoints**
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - User login
- `POST /api/auth/forgot-password` - Password reset request
- `POST /api/auth/reset-password` - Password reset with token

### **Transaction Endpoints**
- `GET /api/transactions` - Get user transactions
- `POST /api/transactions` - Create a new transaction
- `PUT /api/transactions/:id` - Update a transaction
- `DELETE /api/transactions/:id` - Delete a transaction

### **Budget Endpoints**
- `GET /api/budgets` - Get user budgets
- `POST /api/budgets` - Create a new budget
- `PUT /api/budgets/:id` - Update a budget
- `DELETE /api/budgets/:id` - Delete a budget

### **Category Endpoints**
- `GET /api/categories` - Get all categories
- `POST /api/categories` - Create a new category
- `PUT /api/categories/:id` - Update a category
- `DELETE /api/categories/:id` - Delete a category

### **Forecast Endpoints**
- `POST /api/forecast/expenses` - Get expense forecasts

### **Gamification Endpoints**
- `GET /api/gamification/user-status` - Get user's gamification status

---

## **Deployment Guide**

### **AWS Deployment**
1. Set up AWS infrastructure using Terraform scripts in `infrastructure/` directory
2. Configure deployment pipeline in GitHub Actions
3. Deploy backend API to EC2 or ECS
4. Deploy frontend to S3 with CloudFront distribution
5. Set up database in RDS with appropriate security groups
6. Configure monitoring with CloudWatch

### **Docker Deployment**
1. Build backend Docker image:
   ```bash
   docker build -t fintrack-backend ./backend
   ```
2. Build frontend Docker image:
   ```bash
   docker build -t fintrack-frontend ./frontend
   ```
3. Use docker-compose to run the entire stack:
   ```bash
   docker-compose up -d
   ```

---

## **Future Enhancements**
- **Mobile Applications**: Native iOS and Android versions
- **Social Features**: Ability to share savings goals with friends
- **Financial Advisory**: AI-powered financial recommendations
- **Subscription Management**: Track and optimize recurring expenses
- **Multi-Currency Support**: Handle transactions in multiple currencies
- **Data Export**: Export financial data in multiple formats

---

## **Contributors**
- **Backend Team:** Sripriya Dugaputi, Gopinadh Yadlapalli
- **Frontend Team:** Durga Mahesh Boppani, Hemanth Balla

---

---

## **Acknowledgements**
- Special thanks to our mentors and advisors
- All open-source libraries and frameworks used in this project
- The FinTrack team for their dedication and hard work
