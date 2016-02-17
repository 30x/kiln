'use strict'

module.exports = {

  /**
   * Get the env parameter, and set the default if not set
   * @param varName
   * @param defaultValue
   * @returns {*}
     */
  get : function(varName, defaultValue){
    const value = process.env[varName]

    if(!value){
      return defaultValue
    }

    return value
  }
}
