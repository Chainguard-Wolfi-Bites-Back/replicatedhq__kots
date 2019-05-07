import * as React from "react";
import { compose } from "react-apollo";
import { withRouter, Link } from "react-router-dom";
import "../scss/components/Login.scss";

class TraditionalAuth extends React.Component {
  state = {
    email: "",
    password: "",
    firstName: "",
    lastName: ""
  }

  handleLogin = () => {
    console.log("handle login");
  }

  handleSignup = () => {
    console.log("handle sign up");
  }

  onForgotPasswordClick = () => {
    if (this.props.handleForgotPasswordClick && typeof this.props.handleForgotPasswordClick === "function") {
      this.props.handleForgotPasswordClick();
    }
  }

  render() {
    const { context } = this.props;
    const {
      email,
      password,
      firstName,
      lastName
    } = this.state;

    return (
      <div className="flex1 flex-column traditional-auth-wrapper">
        <div className="flex1 flex-column">
          {context === "signup" ?
            <p className="u-lineHeight--normal u-fontSize--large u-color--doveGray u-fontWeight--medium u-marginBottom--20">We just need a few pieces of infomation to get your account created.</p>
          :
            <p className="u-lineHeight--normal u-fontSize--large u-color--doveGray u-fontWeight--medium u-marginBottom--5">Login with your email and password.</p>
          }
          <form>
            {context === "signup" && 
              <div className="u-flexTabletReflow">
                <div className="component-wrapper flex1 u-paddingRight--10">
                  <p className="u-fontSize--normal u-color--tuna u-fontWeight--bold u-lineHeight--normal u-marginBottom--10">First name</p>
                  <input type="text" className="Input" placeholder="John" value={firstName} onChange={(e) => { this.setState({ firstName: e.target.value }) }}/>
                </div>
                <div className="component-wrapper flex1 u-paddingLeft--10">
                  <p className="u-fontSize--normal u-color--tuna u-fontWeight--bold u-lineHeight--normal u-marginBottom--10">Last name</p>
                  <input type="text" className="Input" placeholder="Doe" value={lastName} onChange={(e) => { this.setState({ lastName: e.target.value }) }}/>
                </div>
              </div>
            }
            <div className="flex-column u-marginBottom--10">
              <div className="component-wrapper flex1">
                <p className="u-fontSize--normal u-color--tuna u-fontWeight--bold u-lineHeight--normal u-marginBottom--10">Email address</p>
                <input type="text" className="Input" placeholder="you@example.com" value={email} autoComplete="username" onChange={(e) => { this.setState({ email: e.target.value }) }}/>
              </div>
              <div className="component-wrapper flex1">
                <p className="u-fontSize--normal u-color--tuna u-fontWeight--bold u-lineHeight--normal u-marginBottom--10">Password</p>
                <input type="password" className="Input" placeholder="password" autoComplete="current-password" value={password} onChange={(e) => { this.setState({ password: e.target.value }) }}/>
                {context === "login" && <p className="replicated-link u-fontSize--small u-marginTop--10" onClick={this.onForgotPasswordClick}>Forgot password?</p>}
              </div>
            </div>
          </form>
          <div className="u-marginTop--10 flex alignItems--center">
            <button onClick={context === "signup" ? this.handleSignup : this.handleLogin} className="btn primary">{context === "signup" ? "Create account" : "Log in"}</button>
            <Link to={context === "signup" ? "/login?ta=1" : "/signup"} className="replicated-link u-fontSize--small u-marginLeft--10">{context === "signup" ? "I already have an account" : "Create an account"}</Link>
          </div>
        </div>
      </div>
    );
  }
}

export default compose(
  withRouter
)(TraditionalAuth);
