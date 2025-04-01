const Chart = {
  register: jest.fn(),
  Chart: jest.fn(),
  Pie: jest.fn(),
  Line: jest.fn(),
  Bar: jest.fn(),
  defaults: {
    font: {},
    plugins: {},
  },
};

module.exports = Chart; 