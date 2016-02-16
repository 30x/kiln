'use strict'


module.exports = {


  /**
   * Generate a random int between 0 and 100k
   * @returns {number}
   */

  randomInt: function () {
    return Math.floor(Math.random() * (100000 - 0) + 0);
  }
}



