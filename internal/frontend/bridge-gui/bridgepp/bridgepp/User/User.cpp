// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.


#include "User.h"


namespace bridgepp
{


//****************************************************************************************************************************************************
/// \param[in] parent The parent object of the user.
//****************************************************************************************************************************************************
SPUser User::newUser(QObject *parent)
{
    return SPUser(new User(parent));
}


//****************************************************************************************************************************************************
/// \param[in] parent The parent object.
//****************************************************************************************************************************************************
User::User(QObject *parent)
    : QObject(parent)
{

}


//****************************************************************************************************************************************************
/// \param[in] user The user to copy from
//****************************************************************************************************************************************************
void User::update(User const &user)
{
    this->setID(user.id());
    this->setUsername(user.username());
    this->setPassword(user.password());
    this->setAddresses(user.addresses());
    this->setAvatarText(user.avatarText());
    this->setLoggedIn(user.loggedIn());
    this->setSplitMode(user.splitMode());
    this->setSetupGuideSeen(user.setupGuideSeen());
    this->setUsedBytes(user.usedBytes());
    this->setTotalBytes(user.totalBytes());
}


//****************************************************************************************************************************************************
/// \param[in] makeItActive Should split mode be made active.
//****************************************************************************************************************************************************
void User::toggleSplitMode(bool makeItActive)
{
    emit toggleSplitModeForUser(id_, makeItActive);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::logout()
{
    emit logoutUser(id_);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::remove()
{
    emit removeUser(id_);
}


//****************************************************************************************************************************************************
/// \param[in] address The email address to configure Apple Mail for.
//****************************************************************************************************************************************************
void User::configureAppleMail(QString const &address)
{
    emit configureAppleMailForUser(id_, address);
}


//****************************************************************************************************************************************************
// The only purpose of this call is to forward to the QML application the toggleSplitModeFinished(userID) event
// that was received by the UserList model.
//****************************************************************************************************************************************************
void User::emitToggleSplitModeFinished()
{
    emit toggleSplitModeFinished();
}


//****************************************************************************************************************************************************
/// \return The userID.
//****************************************************************************************************************************************************
QString User::id() const
{
    return id_;
}


//****************************************************************************************************************************************************
/// \param[in] id The userID.
//****************************************************************************************************************************************************
void User::setID(QString const &id)
{
    if (id == id_)
        return;

    id_ = id;
    emit idChanged(id_);
}


//****************************************************************************************************************************************************
/// \return The username.
//****************************************************************************************************************************************************
QString User::username() const
{
    return username_;
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
//****************************************************************************************************************************************************
void User::setUsername(QString const &username)
{
    if (username == username_)
        return;

    username_ = username;
    emit usernameChanged(username_);
}


//****************************************************************************************************************************************************
/// \return The password.
//****************************************************************************************************************************************************
QString User::password() const
{
    return password_;
}


//****************************************************************************************************************************************************
/// \param[in] password The password.
//****************************************************************************************************************************************************
void User::setPassword(QString const &password)
{
    if (password == password_)
        return;

    password_ = password;
    emit passwordChanged(password_);
}


//****************************************************************************************************************************************************
/// \return The addresses.
//****************************************************************************************************************************************************
QStringList User::addresses() const
{
    return addresses_;
}


//****************************************************************************************************************************************************
/// \param[in] addresses The addresses.
//****************************************************************************************************************************************************
void User::setAddresses(QStringList const &addresses)
{
    if (addresses == addresses_)
        return;

    addresses_ = addresses;
    emit addressesChanged(addresses_);
}


//****************************************************************************************************************************************************
/// \return The avatar text.
//****************************************************************************************************************************************************
QString User::avatarText() const
{
    return avatarText_;
}


//****************************************************************************************************************************************************
/// \param[in] avatarText The avatar text.
//****************************************************************************************************************************************************
void User::setAvatarText(QString const &avatarText)
{
    if (avatarText == avatarText_)
        return;

    avatarText_ = avatarText;
    emit usernameChanged(avatarText_);
}


//****************************************************************************************************************************************************
/// \return The login status.
//****************************************************************************************************************************************************
bool User::loggedIn() const
{
    return loggedIn_;
}


//****************************************************************************************************************************************************
/// \param[in] loggedIn The login status.
//****************************************************************************************************************************************************
void User::setLoggedIn(bool loggedIn)
{
    if (loggedIn == loggedIn_)
        return;

    loggedIn_ = loggedIn;
    emit loggedInChanged(loggedIn_);
}


//****************************************************************************************************************************************************
/// \return The split mode status.
//****************************************************************************************************************************************************
bool User::splitMode() const
{
    return splitMode_;
}


//****************************************************************************************************************************************************
/// \param[in] splitMode The split mode status.
//****************************************************************************************************************************************************
void User::setSplitMode(bool splitMode)
{
    if (splitMode == splitMode_)
        return;

    splitMode_ = splitMode;
    emit splitModeChanged(splitMode_);
}


//****************************************************************************************************************************************************
/// \return The 'Setup Guide Seen' status.
//****************************************************************************************************************************************************
bool User::setupGuideSeen() const
{
    return setupGuideSeen_;
}


//****************************************************************************************************************************************************
/// \param[in] setupGuideSeen The 'Setup Guide Seen' status.
//****************************************************************************************************************************************************
void User::setSetupGuideSeen(bool setupGuideSeen)
{
    if (setupGuideSeen == setupGuideSeen_)
        return;

    setupGuideSeen_ = setupGuideSeen;
    emit setupGuideSeenChanged(setupGuideSeen_);
}


//****************************************************************************************************************************************************
/// \return The used bytes.
//****************************************************************************************************************************************************
float User::usedBytes() const
{
    return usedBytes_;
}


//****************************************************************************************************************************************************
/// \param[in] usedBytes The used bytes.
//****************************************************************************************************************************************************
void User::setUsedBytes(float usedBytes)
{
    if (usedBytes == usedBytes_)
        return;

    usedBytes_ = usedBytes;
    emit usedBytesChanged(usedBytes_);
}


//****************************************************************************************************************************************************
/// \return The total bytes.
//****************************************************************************************************************************************************
float User::totalBytes() const
{
    return totalBytes_;
}


//****************************************************************************************************************************************************
/// \param[in] totalBytes The total bytes.
//****************************************************************************************************************************************************
void User::setTotalBytes(float totalBytes)
{
    if (totalBytes == totalBytes_)
        return;

    totalBytes_ = totalBytes;
    emit totalBytesChanged(totalBytes_);
}


} // namespace bridgepp