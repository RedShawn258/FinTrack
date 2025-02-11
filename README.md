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

---

## **Sprint 1 Report**

### **User Stories**

#### **Backend Team**
##### **Sripriya Dugaputi (Signup, Login, API Integration)**
1. **Implement Signup API**
2. **Implement Login API**
3. **Set up authentication routes**
4. **Integrate authentication APIs with backend services**

##### **Gopinadh Yadlapalli (Forgot Password, JWT, Middleware, Database Schema)**
5. **Implement Forgot Password API**
6. **Implement password hashing and JWT token integration**
7. **Set up middleware for authentication & database connection**

#### **Frontend Team**
##### **Durga Mahesh Boppani (Signup, Forgot Password, Context API for State Management)**
8. **Implement Signup Page and Integrate with Backend**
9. **Implement Forgot Password Feature in UI and Connect with Backend**
10. **Enable Context API for State Management**

##### **Hemanth Balla (Login Page, API Integration, Styling & Routing)**
11. **Implement Login Page and Integrate with Backend**
12. **Implement Routing and Navigation for Authentication Pages**
13. **Style Authentication Pages for Better UX**

---

## **Planned Issues for Sprint 1**

1. Backend Authentication APIs (Signup, Login, Forgot Password)
2. Middleware setup for authentication and database connection
3. Frontend Authentication Pages (Signup, Login, Forgot Password)
4. API Integration between frontend and backend
5. State Management with Context API
6. UI Styling and Routing Setup

---

## **Successfully Completed Issues**

✅ Backend Authentication APIs (Signup, Login, Forgot Password)
✅ Middleware setup for authentication and database connection
✅ Frontend Authentication Pages (Signup, Login, Forgot Password)
✅ API Integration between frontend and backend
✅ Context API setup for authentication state management
✅ UI Styling and Routing for authentication pages

---

## **Issues That Were Not Completed & Why**

🚧 **No major pending issues** – all planned tasks were successfully completed in Sprint 1. Some minor refinements and optimizations will be addressed in Sprint 2.

---

## **Next Steps for Sprint 2**
- **Develop Landing Page** with interactive UI.
- **Implement Dashboard & Expense Tracking UI.**
- **Improve UI Components for a more engaging experience.**
- **Expand Backend for advanced financial analytics.**

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

---

## **Contributors**
- **Backend Team:** Sripriya Dugaputi, Gopinadh Yadlapalli
- **Frontend Team:** Durga Mahesh Boppani, Hemanth Balla

