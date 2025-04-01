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

---

## **Sprint 3 Report**

### **User Stories**

#### **Backend Team**
##### **Sripriya Dugaputi (Database & Gamification)**
1. **Implement Secure Database Access Patterns (#52)**
2. **Design Gamification Feature Database Architecture (#53)**
3. **Optimize AWS RDS Database Performance (#51)**

##### **Gopinadh Yadlapalli (Infrastructure & Performance)**
4. **Implement AWS Infrastructure as Code (#56)**
5. **Develop Enhanced Database Monitoring System (#54)**
6. **Conduct Enhanced Load Testing and Performance Optimization (#57)**

#### **Frontend Team**
##### **Durga Mahesh Boppani (Landing Page & Integration)**
7. **Landing Page and Authentication Integration Improvements (#59)**
8. **End-to-End Testing for User Flows (#60)**
9. **Refine "Read More" Information Pages (#58)**

##### **Hemanth Balla (UI Testing & Refinement)**
10. **Landing Page Component Unit Tests (#62)**
11. **Information Page Unit Tests (#63)**
12. **Landing Page Content Refinement (#61)**

### **Successfully Completed Issues**

✅ Landing Page Components and Authentication Integration
✅ Information Pages with "Read More" Sections
✅ Comprehensive Testing Strategy (Unit Tests and E2E)
✅ AWS Infrastructure with IaC Principles
✅ Database Performance Monitoring and Optimization
✅ Secure Database Access Patterns
✅ Gamification Feature Architecture
✅ Load Testing and Performance Optimization

---

## **Sprint 2 Report**

### **User Stories**

#### **Backend Team**
##### **Sripriya Dugaputi (Budget & Categories)**
1. **Implement Budget management and Category APIs**
2. **Create Unit Tests for the respective APIs**
3. **Build Database tables and respective schema**

##### **Gopinadh Yadlapalli (Transaction API & Testing)**
4. **Implement Transaction API and Develop unit test cases for it**
5. **Create Unit Tests for User Authentication**
6. **Implement API routings for the new APIs**

#### **Frontend Team**
##### **Durga Mahesh Boppani (Dashbaord & UI)**
7. **Implement Dashboard UI with styling**
8. **Create Data Visualization Components**
9. **Write Unit Tests for Dashboard feature**

##### **Hemanth Balla (Testing & API calls)**
10. **Implement Cypress End-to-End Tests**
11. **Integrate API calls to the backend server**
12. **Create Unit Tests for User Authentication**

### **Successfully Completed Issues**

✅ Expense Tracking Implementation
✅ Budget Management Features
✅ Data Visualization Components
✅ Responsive Design Implementation
✅ Comprehensive Unit Tests
✅ End-to-End Testing with Cypress
✅ API Documentation
✅ Error Handling and Validation

---

## **Next Steps for Sprint 4**
- **Implement User Profile Customization**
- **Develop Social Sharing Features**
- **Add Expense Forecasting Capabilities**
- **Refine Mobile App Experience**
- **Improve Accessibility Features**

---

## **Setup & Installation**

### **Backend Setup**
1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/FinTrack.git
   cd FinTrack/backend
   ```
2. Install dependencies:
   ```bash
   go mod init github.com/RedShawn258/FinTrack/backend
   go mod tidy
   ```
3. Start the server:
   ```bash
   cd backend
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
3. Start the frontend application:
   ```bash
   npm start
   ```

### **Running Tests**
1. Frontend Tests:
   ```bash
   npm test
   ```
2. Cypress E2E Tests:
   ```bash
   npm run cypress:open
   ```
3. Backend Tests:
   ```bash
   go test ./...
   ```

---

## **Contributors**
- **Backend Team:** Sripriya Dugaputi, Gopinadh Yadlapalli
- **Frontend Team:** Durga Mahesh Boppani, Hemanth Balla



